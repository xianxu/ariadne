# Release template: scripts/release-weave.sh fills metadata from its archives.
class Weave < Formula
  desc "Prepare repository layers and compile agent context"
  homepage "https://github.com/xianxu/ariadne"
  version "@WEAVE_VERSION@"

  on_macos do
    on_arm do
      url "@WEAVE_DARWIN_ARM64_URL@"
      sha256 "@WEAVE_DARWIN_ARM64_SHA256@"
    end
    on_intel do
      url "@WEAVE_DARWIN_AMD64_URL@"
      sha256 "@WEAVE_DARWIN_AMD64_SHA256@"
    end
  end
  on_linux do
    on_arm do
      url "@WEAVE_LINUX_ARM64_URL@"
      sha256 "@WEAVE_LINUX_ARM64_SHA256@"
    end
    on_intel do
      url "@WEAVE_LINUX_AMD64_URL@"
      sha256 "@WEAVE_LINUX_AMD64_SHA256@"
    end
  end

  def install
    bin.install "weave"
  end

  test do
    assert_match "weave version #{version}", shell_output("#{bin}/weave --version")
    (testpath/"base/construct").mkpath
    (testpath/"base/construct/base.manifest").write "export prose AGENTS.base.md\n"
    (testpath/"base/AGENTS.base.md").write "Packaged weave composes local layers.\n"
    (testpath/"leaf").mkpath
    cd testpath/"leaf" do
      system bin/"weave", "link", "../base"
      system bin/"weave", "compile"
      assert_match "Packaged weave composes local layers.", (testpath/"leaf/AGENTS.md").read
    end
  end
end
