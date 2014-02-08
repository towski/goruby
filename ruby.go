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

type Declaration struct {
    array []int
    file string
}

func(d *Declaration) Unshift() int {
    current_line, d.array = d.array[len(d.array)-1], d.array[:len(d.array)-1]
    return current_line
}

func(d *Declaration) Push(line_number int) {
    d.array = append(d.array, 0)
    copy(d.array[1:], d.array[0:])
    d.array[0] = line_number
}

func NewObject(class *Object) Object {
    o := Object{}
    o.class = class
    if(o.class == CLASS){
        o.methods = map[string]func(...interface{}) *Object {}
    }
    return o
}

func (o *Object) GetMethod(method string) func(...interface{})*Object {
      fmt.Printf("looking for method %s on %s\n", method, o.class)
    if(o.class == CLASS){
      fmt.Printf("looking in here\n")
        meth, ok := o.methods[method]
        if(ok){
            return meth
        } else {
      fmt.Printf("looking in here\n")
            return KERNEL.methods[method]
        }
    } else {
        return o.class.GetMethod(method)
    }
}

var definition_locations map[string] *Declaration
var classes map[string]*Object
var constants map[string]Object
var locals map[string]*Object
var current_line int
var args []Object
var first_regex *regexp.Regexp
var second_regex *regexp.Regexp
var bytecodeMap map[string]func(...interface{})
var call_object *Object
var last_return *Object
var CLASS *Object
var IO *Object
var OBJECT *Object
var KERNEL *Object
var STRING *Object
var Nil *Object

func putobject(v ...interface{}){
    args = append(args, NewObject(OBJECT))
}

func putself(v ...interface{}){
    //call_object = "a"
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

func putspecialobject(v ...interface{}){
  fmt.Printf("%s", v[0].(string))
}

func defineclass(v ...interface{}){
  var arr []string = strings.Split(v[0].(string), ",")
  fmt.Printf("%s\n", arr[1])
  definition := definition_locations[strings.Trim(arr[1], " ")]
  fmt.Printf("%d\n", definition.array)
  call_object := NewObject(CLASS)
  classes[strings.Trim(arr[0], " ")] = &call_object
  current_line = definition.Unshift()
  fmt.Printf("%d\n", current_line)
}

func step(line string) {
    var starting_number = current_line
    fmt.Printf("%s\n", line)
    if(!first_regex.MatchString(line)){
        var match = second_regex.FindStringSubmatch(line)
        if(len(match) > 1){
            var arguments = strings.Trim(match[2], " ")
            bytecodeMap[match[1]](arguments)
        }
    }
    if(starting_number == current_line){
        current_line += 1
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
    var kernel Object = NewObject(CLASS)
    KERNEL = &kernel
    KERNEL.methods[":puts"] = func(v ...interface{}) *Object {
        fmt.Printf(v[0].([]Object)[0].data)
        return Nil
    }
    KERNEL.methods[":\"core#define_method\""] = func(v ...interface{}) *Object {
        fmt.Printf("define method")
        return Nil
    }
    call_object = KERNEL
    classes[":IO"] = IO
    classes[":Kernel"] = KERNEL
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
            "putspecialobject": putspecialobject,
            "defineclass": defineclass,
            "pop": putnil,
            "putself": putself,
            "putiseq": putnil,
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
    definition_locations = map[string]*Declaration{}
    var lines []string
    var line string
    var last_class string
    var line_number int = 0
    definition_regex, _ := regexp.Compile(`== disasm: <RubyVM::InstructionSequence:([^@]*)@([^\>]*)>=*`)
    for scanner.Scan() {
        line = scanner.Text()
        var match = definition_regex.FindStringSubmatch(line)
        if(len(match) > 1){
            var key string
            if match[1][0] == '<' {
                last_class = match[1]
                key = match[1]
            } else {
                key = last_class + match[1]
            }
            declaration, present := definition_locations[key]
            if(!present){
                declaration = &Declaration {}
                definition_locations[key] = declaration
            }
            declaration.Push(line_number)
            fmt.Printf("%s\n", key)
            // TODO: handle second file argument
        }
        lines = append(lines, line)
        line_number += 1
    }
    // TODO: multiple definition locations
    main := definition_locations["<main>"]
    current_line = main.Unshift()
    for(true){
        step(lines[current_line])
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
