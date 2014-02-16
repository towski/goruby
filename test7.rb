class Hey
  def hey &block
    a = yield
    puts "a"
    a
  end
end
obj = Hey.new
puts(obj.hey { "this" })
