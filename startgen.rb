class Startgen < Formula
  desc "A CLI tool to generate boilerplate projects using natural language"
  homepage "https://github.com/mario2334/startgen"
  url "https://github.com/mario2334/startgen/releases/download/v0.1.1-alpha/startgen-darwin-amd64.zip" # Update for your release
  sha256 "19f5e0da672f8cfb4684a2cb68e8b929e56f3f573abaf903121eea5fec452f84" # Calculate using 'shasum -a 256 <file>'
  version "1.0.0"

  def install
    bin.install "startgen"
  end

  test do
    assert_match "Usage", shell_output("#{bin}/startgen --help", 2)
  end
end