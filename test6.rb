c = Class.new do
  def hey
    puts "a"
    [[]].each do |a|
      a.each do |b|
        puts "c"
      end
    end
  end
end
obj = c.new
obj.hey
