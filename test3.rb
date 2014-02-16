class Jumbo
  def initialize
    @file = 'hey'
  end

  def main 
    ["hey"].each do |i|
      puts i
    end

    ["hey"].each do |i|
      puts i
    end
  end
end

jumbo = Jumbo.new
jumbo.main :this
