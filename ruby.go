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
    class_name string
    methods map[string]func(...*Object) *Object
    data string
    definition_index int
}

type Declaration struct {
    array []int
    file string
}

type Context struct {
    definition map[string]*ByteCode
    line_number string
}

type Stack struct {
    array []Context
}

type ByteCode struct {
    code string
    params string
    line_number string
    next_code *ByteCode
}

func NewByteCode(line string) (*ByteCode, string) {
    var match = second_regex.FindStringSubmatch(line)
    b := ByteCode{}
    var line_number string
    if(len(match) > 2){
        line_number = match[1]
        b.line_number = line_number
        b.code = match[2]
        b.params = strings.Trim(match[3], " ")
    }
    return &b, line_number
}

func(d *Declaration) Unshift() int {
    var current_line int
    current_line, d.array = d.array[len(d.array)-1], d.array[:len(d.array)-1]
    return current_line
}

func(d *Declaration) Shift(line_number int) {
    d.array = append(d.array, 0)
    copy(d.array[1:], d.array[0:])
    d.array[0] = line_number
}

func(s *Stack) Pop() (map[string]*ByteCode, string) {
    var context Context = s.array[len(s.array)-1]
    s.array = s.array[:len(s.array)-1]
    return context.definition, context.line_number
}

func(s *Stack) Push(definition map[string]*ByteCode, line_number string)  {
    s.array = append(s.array, Context{definition:definition, line_number:line_number})
}

func NewObject(class *Object, name string) *Object {
    o := Object{}
    o.class = class
    o.class_name = name
    if(o.class == CLASS){
        o.methods = map[string]func(...*Object) *Object {}
    }
    return &o
}

func (o *Object) GetMethod(method string, class_method bool) (func(...*Object)*Object, bool) {
    meth, ok := o.methods[method]
    if(ok){
        return meth, true
    } else {
        // append ourself to the call arguments if we're calling a class method
        args = append(args, nil)
        copy(args[1:], args[0:])
        args[0] = o
        meth, ok2 := o.class.methods[method]
        if(ok2){
            return meth, true
        } else {
            meth, ok := KERNEL.methods[method]
            return meth, ok
        }
    }
}

var definition_locations map[string][]map[string]*ByteCode
var classes map[string]*Object
var constants map[string]Object
var locals map[string]*Object
var current_line_number string
var args []*Object
var first_regex *regexp.Regexp
var second_regex *regexp.Regexp
var bytecodeMap map[string]func(...interface{})
var scope *Object
var last_return *Object
var CLASS *Object
var IO *Object
var OBJECT *Object
var KERNEL *Object
var STRING *Object
var Nil *Object
var stack Stack
var last_call_name string
var current_definition_location map[string]*ByteCode
var current_definition_key string

func putobject(v ...interface{}){
    var obj = NewObject(STRING, "")
    obj.data = v[0].(string)
    args = append(args, obj)
}

func putself(v ...interface{}){
    //scope = "a"
}

func putstring(v ...interface{}){
    var obj = NewObject(STRING, "")
    obj.data = v[0].(string)
    args = append(args, obj)
}

func putnil(v ...interface{}){ }

func getconstant(v ...interface{}){
    scope = classes[v[0].(string)]
    if(false){
        fmt.Printf("Constant not defined in scope")
        os.Exit(1)
    }
}

func send(v ...interface{}){
    var arr []string = strings.Split(v[0].(string), ",")
    function, ok := scope.GetMethod(arr[0], false)
    if(ok){
        last_call_name = arr[0]
        last_return = function(args...)
        args = []*Object{}
    } else {
        fmt.Printf("Function not defined in scope")
        os.Exit(1)
    }
}

func setlocal(v ...interface{}){
    locals[v[0].(string)] = last_return
}

func getlocal(v ...interface{}){
    scope = locals[v[0].(string)]
}

func leave(v ...interface{}){
    if(len(stack.array) != 0){
        current_definition_location, current_line_number = stack.Pop()
        scope = KERNEL
    } else {
        os.Exit(0)
    }
}

func putspecialobject(v ...interface{}){
}

func defineclass(v ...interface{}){
    var arr []string = strings.Split(v[0].(string), ",")
    var location_name string = strings.Trim(arr[1], " ")
    var name string = strings.Trim(arr[0], " ")
    var class *Object
    // right now we expect that multiple class definition locations are in the order that they're called
    definition := definition_locations[location_name]
    class, ok := classes[name]
    if(ok){
        scope = class
    } else {
        scope = NewObject(CLASS, name)
        classes[name] = scope
    }
    stack.Push(current_definition_location, current_definition_location[current_line_number].next_code.line_number)
    // should be unshift not pop
    current_definition_location = definition[scope.definition_index]
    current_line_number = "0000"
    scope.definition_index += 1
}

func step(code *ByteCode) {
    var starting_number = current_line_number
    fmt.Printf("code  %s %s %s\n", code.line_number, code.code, code.params)
    bytecodeMap[code.code](code.params)
    if(starting_number == current_line_number){
        current_line_number = code.next_code.line_number
        step(code.next_code)
    } else {
        step(current_definition_location[current_line_number])
    }
}

func setup(){
    stack = Stack{}
    CLASS = &Object{}
    CLASS.methods = map[string]func(...*Object) *Object {}
    CLASS.methods[":new"] = func(v ...*Object) *Object {
        var class *Object = v[0]
        var obj = NewObject(class, "")
        return obj
    }
    classes = map[string]*Object {}
    locals = map[string]*Object {}
    Nil = NewObject(CLASS, "Nil")
    IO = NewObject(CLASS, "IO")
    IO.methods[":new"] = func(v ...*Object) *Object {
        var obj = NewObject(IO, "IO")
        return obj
    }
    IO.methods[":puts"] = func(v ...*Object) *Object {
        fmt.Printf("puts %s\n", v[0].data)
        return Nil
    }
    KERNEL = NewObject(CLASS, "Kernel")
    KERNEL.methods[":puts"] = func(v ...*Object) *Object {
        if(len(v) == 1){
            fmt.Printf("%s\n", v[0].data)
        } else {
            fmt.Printf("%s\n", v[1].data)
        }
        return Nil
    }
    KERNEL.methods[":\"core#define_method\""] = func(v ...*Object) *Object {
        var name string = v[1].data
        scope.methods[name] = func(v ...*Object) *Object {
            var class_name string = v[0].class.class_name
            var key = "<class" + class_name + ">" + last_call_name
            stack.Push(current_definition_location, current_line_number)
            current_line_number = "0000"
            current_definition_location = definition_locations[key][0]
            //definition.Unshift()
            return Nil
        }
        return Nil
    }
    scope = KERNEL
    classes[":IO"] = IO
    classes[":Kernel"] = KERNEL
    first_regex, _ = regexp.Compile(`==`)
    second_regex, _ = regexp.Compile(`(^\d+) ([^\(\s]*)([^\(]*)(\(.*){0,1}`)
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
    definition_locations = map[string][]map[string]*ByteCode{}
    current_definition_location = map[string]*ByteCode{}
    var temp_definition_location map[string]*ByteCode
    var line string
    var last_class string
    var last_byte_code *ByteCode
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
                // last_call_name is a :symbol, so definition keys have a : added
                key = last_class + ":" + match[1]
            }
            temp_definition_location = map[string]*ByteCode{}
            definition_locations[key] = append(definition_locations[key], temp_definition_location)
            // TODO: handle second file argument
        }
        byte_code, line_number := NewByteCode(line)
        if(last_byte_code != nil){
            last_byte_code.next_code = byte_code
        }
        if(line_number != ""){
            temp_definition_location[line_number] = byte_code
            last_byte_code = byte_code
        }
        //lines = append(lines, line)
        //line_number += 1
    }
    // TODO: multiple definition locations
    current_definition_location = definition_locations["<main>"][0]
    current_line_number = "0000"
                fmt.Println(current_definition_location[current_line_number].code)
    step(current_definition_location[current_line_number])
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
