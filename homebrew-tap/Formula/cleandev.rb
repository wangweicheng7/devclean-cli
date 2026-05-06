class Cleandev < Formula
  desc "macOS developer cleanup CLI (safe-first)"
  homepage "https://github.com/wangweicheng7/cleandev-cli"
  version "0.0.0"

  # Current default: build from GitHub main tarball.
  # For releases, update url/sha256/version to the corresponding tag tarball.
  url "https://github.com/wangweicheng7/cleandev-cli/archive/refs/heads/main.tar.gz"
  sha256 "59223bbf969eb095d800e6eac7bef524088b9aae40b6053824af689ce7a4d3ce"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(output: bin/"cleandev"), "./cmd/cleandev"
  end

  test do
    system "#{bin}/cleandev", "doctor"
  end
end

