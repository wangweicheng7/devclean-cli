class DevcleanCli < Formula
  desc "macOS developer cleanup CLI (safe-first)"
  homepage "https://github.com/wangweicheng7/devclean-cli"
  version "v0.1.0"

  # For releases, update url/sha256/version to the corresponding tag tarball.
  url "https://github.com/wangweicheng7/devclean-cli/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "12668827a2ccb9fdb2c05087588b78133c252c6e14bb827ec1af57c8f5e211ce"
  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(output: bin/"devclean"), "./cmd/cleandev"
  end

  test do
    system "#{bin}/devclean", "doctor"
  end
end

