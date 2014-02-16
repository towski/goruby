class Fish
  def hey a, &block
    puts "a" do
      puts "d"
    end
    yield
    puts "a"
  end
end
obj = Fish.new
obj.hey do
  puts "b"
end
