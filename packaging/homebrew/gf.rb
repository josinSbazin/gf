# typed: false
# frozen_string_literal: true

class Gf < Formula
  desc "CLI for GitFlic - Russian GitHub alternative"
  homepage "https://github.com/josinSbazin/gf"
  version "0.3.0"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/josinSbazin/gf/releases/download/v0.3.0/gf-darwin-amd64.tar.gz"
      sha256 "cd27adc55817bd9796226d6b3893f300578010ebf250a4f7eee7de72d491bc8d"

      def install
        bin.install "gf-darwin-amd64" => "gf"
      end
    end

    on_arm do
      url "https://github.com/josinSbazin/gf/releases/download/v0.3.0/gf-darwin-arm64.tar.gz"
      sha256 "b59aeb4a836dd4cd8d47824c85c2f83ecbeee5812bd356cbae50149c249d6e80"

      def install
        bin.install "gf-darwin-arm64" => "gf"
      end
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/josinSbazin/gf/releases/download/v0.3.0/gf-linux-amd64.tar.gz"
      sha256 "67e8db072b8837ab9b83b24edb32e520d41592abe3e07bac61161ec9e4fb0910"

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
