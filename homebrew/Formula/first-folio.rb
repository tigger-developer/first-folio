# ABOUTME: Homebrew formula for First Folio.
# ABOUTME: Copy this file to tigger-developer/homebrew-tap/Formula/ after tagging a release.

class FirstFolio < Formula
  desc "Convert stage plays to Org, Markdown, Fountain, PDF, DOCX and ODT"
  homepage "https://github.com/tigger-developer/first-folio"
  url "https://github.com/tigger-developer/first-folio/archive/refs/tags/v0.4.16.tar.gz"
  sha256 "e50633fbe8979a8f5c438c1caf4f9422cd72ab6cf140a0f45ca7d99e13b6c74e"
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
      Fountain, and PDF formats, with editable DOCX and ODT export.
      PDF output requires Typst. ODT output requires Pandoc.

      Try it:
        folio convert play.org play.pdf
        folio convert play.org play.odt
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
