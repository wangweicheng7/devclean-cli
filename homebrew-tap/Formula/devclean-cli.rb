class DevcleanCli < Formula
  desc "macOS developer cleanup CLI (safe-first)"
  homepage "https://github.com/wangweicheng7/devclean-cli"
  version "v0.1.0"

  # For releases, update url/sha256/version to the corresponding tag tarball.
  url "https://github.com/wangweicheng7/devclean-cli/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "ac19f3fca77921d11b6e6cc0314ff407607c3ea600131fe756debcdcc3c94fc4"
  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(output: bin/"devclean"), "./cmd/devclean"
  end

  test do
    system "#{bin}/devclean", "doctor"
  end
end

