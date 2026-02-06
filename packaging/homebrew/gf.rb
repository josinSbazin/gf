# typed: false
# frozen_string_literal: true

class Gf < Formula
  desc "CLI for GitFlic - Russian GitHub alternative"
  homepage "https://github.com/josinSbazin/gf"
  version "0.2.0"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/josinSbazin/gf/releases/download/v0.2.0/gf-darwin-amd64.tar.gz"
      sha256 "57caad89db23937f59c7d82fceb2291cb72f60a77ac7f7496329193b09af76f2"

      def install
        bin.install "gf-darwin-amd64" => "gf"
      end
    end

    on_arm do
      url "https://github.com/josinSbazin/gf/releases/download/v0.2.0/gf-darwin-arm64.tar.gz"
      sha256 "088ee35a379bcea96bef81e766780b89a3fd13b57c31f2cade86e9becaa30e9f"

      def install
        bin.install "gf-darwin-arm64" => "gf"
      end
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/josinSbazin/gf/releases/download/v0.2.0/gf-linux-amd64.tar.gz"
      sha256 "e568105698c1c3dacc72d50eba5c4019fe42651fce0922b9ea5b44c7a3c5fcd0"

      def install
        bin.install "gf-linux-amd64" => "gf"
      end
    end
  end

  def caveats
    <<~EOS
      To get started, run:
        gf auth login

      Get your API token at: https://gitflic.ru/settings/oauth/token
    EOS
  end

  test do
    assert_match "gf version", shell_output("#{bin}/gf --version", 0)
  end
end
