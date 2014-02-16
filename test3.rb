class Fish
  def hey a, &block
    yield
    puts "a"
  end
end
obj = Fish.new
obj.hey do
  puts "b"
end
