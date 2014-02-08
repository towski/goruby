package main

import("fmt")
import("os/exec")
import("os")
import("sync")
import("strings")
import("bufio")
import("regexp")

type Object struct {
    class *Object
    name string
    methods map[string]func(...interface{}) *Object
    data string
}

func NewObject(class *Object) Object {
    o := Object{}
    o.class = class
    o.methods = map[string]func(...interface{}) *Object {}
    return o
}

func (o *Object) GetMethod(method string) func(...interface{})*Object {
    if(o.class == CLASS){
        return o.methods[method]
    } else {
        return o.class.GetMethod(method)
    }
}

var classes map[string]*Object
var constants map[string]Object
var locals map[string]*Object
var args []Object
var first_regex *regexp.Regexp
var second_regex *regexp.Regexp
var bytecodeMap map[string]func(...interface{})
var call_object *Object
var last_return *Object
var CLASS *Object
var IO *Object
var OBJECT *Object
var STRING *Object
var Nil *Object

func putobject(v ...interface{}){
    args = append(args, NewObject(OBJECT))
}

func putstring(v ...interface{}){
    var obj = NewObject(STRING)
    obj.data = v[0].(string)
    args = append(args, obj)
}

func putnil(v ...interface{}){ }

func getconstant(v ...interface{}){
    call_object = classes[v[0].(string)]
    if(false){
        os.Exit(1)
    }
}

func send(v ...interface{}){
    var arr []string = strings.Split(v[0].(string), ",")
    last_return = call_object.GetMethod(arr[0])(args)
    args = []Object{}
}

func setlocal(v ...interface{}){
    locals[v[0].(string)] = last_return
}

func getlocal(v ...interface{}){
    call_object = locals[v[0].(string)]
}

func leave(v ...interface{}){
    os.Exit(0)
}

func step(line string, line_number int) {
    if(!first_regex.MatchString(line)){
        var match = second_regex.FindStringSubmatch(line)
        if(len(match) > 1){
            var arguments = strings.Trim(match[2], " ")
            bytecodeMap[match[1]](arguments)
        }
    }
}

func setup(){
    CLASS = &Object{}
    classes = map[string]*Object {}
    locals = map[string]*Object {}
    var nil Object = NewObject(CLASS)
    Nil = &nil
    var io Object = NewObject(CLASS)
    IO = &io
    IO.methods[":new"] = func(v ...interface{}) *Object {
        var obj = NewObject(IO)
        return &obj
    }
    IO.methods[":puts"] = func(v ...interface{}) *Object {
        fmt.Printf(v[0].([]Object)[0].data)
        return Nil
    }
    classes[":IO"] = IO
    first_regex, _ = regexp.Compile(`==`)
    second_regex, _ = regexp.Compile(`^\d+ ([^\(\s]*)([^\(]*)(\(.*){0,1}`)
    bytecodeMap = map[string]func(...interface{}) {
            "putobject": putobject,
            "putstring": putstring,
            "putnil": putnil,
            "getconstant": getconstant,
            "send": send,
            "setlocal": setlocal,
            "getlocal": getlocal,
            "leave": leave,
    }
}

func execute_cmd(cmd string, wg *sync.WaitGroup) {
    setup()
    parts := strings.Fields(cmd)
    head := parts[0]
    parts = parts[1:len(parts)]

    out, err := exec.Command(head,parts...).Output()
    if err != nil {
      fmt.Printf("%s", err)
    }
    scanner := bufio.NewScanner(strings.NewReader(string(out)))
    var line_number int = 0
    for scanner.Scan() {
      step(scanner.Text(), line_number) 
    }
    wg.Done() 
}

func main() {
    wg := new(sync.WaitGroup)
    wg.Add(1)
    if(len(os.Args) == 1){
        fmt.Printf("Requires file input argument (ex: test.rb)\n")
        os.Exit(1)
    }
    execute_cmd("disassemble.rb " + os.Args[1], wg)
    wg.Wait()
}
