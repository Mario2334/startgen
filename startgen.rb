class Startgen < Formula
  desc "A CLI tool to generate boilerplate projects using natural language"
  homepage "https://github.com/mario2334/startgen"
  url "https://github.com/Mario2334/startgen/releases/download/v0.1.1-alpha/startgen-darwin-amd64.zip" # Update for your release
  sha256 "be17add25d763c22bd8e1487406b3e2fa4ee61f61cd23199f9d7360478d4f375" # Calculate using 'shasum -a 256 <file>'
  version "0.1.1-alpha"

  def install
    bin.install "startgen"
  end

  test do
    assert_match "Usage", shell_output("#{bin}/startgen --help", 2)
  end
end