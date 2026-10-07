# ABOUTME: Homebrew formula for First Folio.
# ABOUTME: Copy this file to tigger-developer/homebrew-tap/Formula/ after tagging a release.

class FirstFolio < Formula
  desc "Format converter for stage plays — org, markdown, fountain, PDF"
  homepage "https://github.com/tigger-developer/first-folio"
  url "https://github.com/tigger-developer/first-folio/archive/refs/tags/v0.4.14.tar.gz"
  sha256 "23ed2756c86e2a60b02dd30a473c41dae287ab9a7b608aa938765ecd0d9fb6f4"
  license "MIT"
  head "https://github.com/tigger-developer/first-folio.git", branch: "master"

  depends_on "typst"
  depends_on "pandoc"
  depends_on "go" => :build

  def install
    system "go", "build", "-trimpath",
           "-ldflags", "-X github.com/tigger-developer/first-folio/internal/app.Version=#{version}",
           "-o", bin/"folio", "./cmd/folio"
  end

  def caveats
    <<~EOS
      First Folio converts stage plays between org-mode, Markdown,
      Fountain, and PDF formats. PDF output requires Typst.

      Try it:
        folio convert play.org play.pdf
        folio convert play.org --to md
        folio letter play.org

      Config: ~/.config/first-folio/script.yaml
      Styles: --style=british (default), --style=us, --style=screenplay

      See: #{homepage}
    EOS
  end

  test do
    assert_match "folio", shell_output("#{bin}/folio --version")
    assert_match "convert", shell_output("#{bin}/folio --help")
  end
end
