class DevcleanCli < Formula
  desc "macOS developer cleanup CLI (safe-first)"
  homepage "https://github.com/wangweicheng7/devclean-cli"
  version "0.2.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/wangweicheng7/devclean-cli/releases/download/v0.2.0/devclean_0.2.0_darwin_arm64.tar.gz"
      sha256 "38acd9c79e5c1ef055a2f8479c0ae564ae3ae592618ec7ab8121ba288f01a2d7"
    else
      url "https://github.com/wangweicheng7/devclean-cli/releases/download/v0.2.0/devclean_0.2.0_darwin_amd64.tar.gz"
      sha256 "c645d4d0a4080de48ff7328a32dffaf7be72a171e2b2760ae2f04234169badec"
    end
  end

  def install
    bin.install "devclean"
  end

  test do
    system "#{bin}/devclean", "doctor"
  end
end

