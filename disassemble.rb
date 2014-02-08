#!/usr/bin/env ruby
file = ARGV[0]
seq = RubyVM::InstructionSequence.compile_file(file, false)
puts seq.disassemble
