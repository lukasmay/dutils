class Dutils < Formula
  desc "Docker and Docker Compose workflow simplifier"
  homepage "https://github.com/lukasmay/dutils"
  url "https://github.com/lukasmay/dutils/archive/refs/tags/v1.0.0.tar.gz"
  sha256 "" # This would be filled in when you release
  license "Apache-2.0"

  depends_on "go" => :build

  def install
    system "go", "build", "-ldflags", "-s -w -X github.com/lukasmay/dutils/cmd.Version=1.0.0", "-o", bin/"dutils", "main.go"
    
    # Generate shell completions
    (bash_completion/"dutils").write `#{bin}/dutils completion bash`
    (zsh_completion/"_dutils").write `#{bin}/dutils completion zsh`
    (fish_completion/"dutils.fish").write `#{bin}/dutils completion fish`
  end

  test do
    assert_match "dutils 1.0.0", shell_output("#{bin}/dutils version")
  end
end
