c = Class.new do
  def hey &block
    puts "a"
    as = yield
    puts "b"
  end
end
obj = c.new
obj.hey
