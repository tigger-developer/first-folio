# ABOUTME: Prose manuscript renderer for Markdown and org-mode chapter files.
# ABOUTME: Resolves chapter inputs, parses prose structure, and emits Typst/PDF.
package Folio::Manuscript;

use strict;
use warnings;
use utf8;

use File::Basename qw(dirname);
use File::Temp qw(tempfile);

sub render {
    my ($class, %opts) = @_;

    my $inputs = $opts{inputs} || [];
    my $output = $opts{output} or die "Error: no manuscript output specified\n";
    my $config = $opts{config} or die "Error: manuscript render requires config\n";

    my @paths = _resolve_inputs(@$inputs);
    my $format = _validate_input_formats(@paths);
    _validate_text_files(@paths);
    my $source = _read_joined_files(@paths);
    my $doc = $format eq 'markdown' ? _parse_markdown($source) : _parse_org($source);

    _apply_config_metadata($doc, $config, $opts{metadata} || {});

    my $typst = _render_typst($doc, $config);
    if ($output =~ /\.typ$/i) {
        _write_text($output, $typst);
        return $output;
    }
    if ($output !~ /\.pdf$/i) {
        die "Error: manuscript output must end in .typ or .pdf: $output\n";
    }

    my ($tmp_fh, $tmp_path) = tempfile('folio-manuscript-XXXX', SUFFIX => '.typ', TMPDIR => 1);
    binmode($tmp_fh, ':encoding(UTF-8)');
    print $tmp_fh $typst;
    close $tmp_fh;

    my @cmd = ('typst', 'compile', $tmp_path, $output);
    if (system(@cmd) != 0) {
        my $exit = $?;
        unlink $tmp_path if -e $tmp_path;
        die "Error: typst compile failed (exit $exit)\n";
    }

    unlink $tmp_path;
    return $output;
}

sub _resolve_inputs {
    my (@inputs) = @_;
    die "Error: no manuscript input files specified\n" if !@inputs;

    my %seen;
    my @paths;
    for my $input (@inputs) {
        my @matches = glob($input);
        @matches = ($input) if !@matches;
        for my $path (@matches) {
            if (!-f $path) {
                die "Error: manuscript input not found: $path\n";
            }
            next if $seen{$path};
            $seen{$path} = 1;
            push @paths, $path;
        }
    }

    @paths = sort @paths;
    return @paths;
}

sub _validate_input_formats {
    my (@paths) = @_;

    my %formats;
    for my $path (@paths) {
        my $format;
        if ($path =~ /\.md$/i) {
            $format = 'markdown';
        } elsif ($path =~ /\.org$/i) {
            $format = 'org';
        } elsif ($path =~ /\.(?:fountain|ftn)$/i) {
            die "Error: manuscript mode accepts only Markdown or org-mode, not Fountain: $path\n";
        } else {
            die "Error: manuscript mode accepts only Markdown or org-mode input: $path\n";
        }
        $formats{$format} = 1;
    }

    if (keys(%formats) > 1) {
        die "Error: manuscript mode cannot mix Markdown and org-mode input files\n";
    }

    return (keys %formats)[0];
}

sub _validate_text_files {
    my (@paths) = @_;

    for my $path (@paths) {
        open(my $raw_fh, '<:raw', $path)
            or die "Error: cannot open $path: $!\n";
        my $buf;
        read($raw_fh, $buf, 8192);
        close $raw_fh;

        if ($buf =~ /\x00/) {
            die "Error: $path appears to be a binary file, not a text document.\n";
        }

        open(my $utf8_fh, '<:encoding(UTF-8)', $path)
            or die "Error: cannot open $path as UTF-8: $!\n";
        eval {
            local $SIG{__WARN__} = sub {
                die "encoding warning: $_[0]";
            };
            while (<$utf8_fh>) {
                # read through to trigger decoding warnings
            }
        };
        close $utf8_fh;
        if ($@) {
            die "Error: $path has invalid encoding: $@\n";
        }
    }
}

sub _read_joined_files {
    my (@paths) = @_;

    my @chunks;
    for my $path (@paths) {
        open(my $fh, '<:encoding(UTF-8)', $path)
            or die "Error: cannot open $path: $!\n";
        local $/;
        my $text = <$fh>;
        close $fh;
        push @chunks, $text;
    }

    return join("\n\n", map { _trim_trailing_newlines($_) } @chunks);
}

sub _trim_trailing_newlines {
    my ($text) = @_;
    $text =~ s/\s+\z//;
    return $text;
}

sub _new_doc {
    return {
        meta     => {},
        elements => [],
        toc      => [],
        footnote => {},
    };
}

sub _parse_markdown {
    my ($source) = @_;
    my $doc = _new_doc();
    my @lines = split /\n/, $source;
    my $skip_level = 0;
    my $in_code = 0;
    my @paragraph;

    my $flush = sub {
        return if !@paragraph;
        push @{$doc->{elements}}, { type => 'paragraph', text => join("\n", @paragraph) };
        @paragraph = ();
    };

    for my $line (@lines) {
        if ($line =~ /^```/) {
            $flush->();
            $in_code = !$in_code;
            push @{$doc->{elements}}, { type => $in_code ? 'code_start' : 'code_end' };
            next;
        }
        if ($in_code) {
            push @{$doc->{elements}}, { type => 'code', text => $line };
            next;
        }

        if ($line =~ /^(#{1,6})\s+(.+)$/) {
            $flush->();
            my $level = length($1);
            my $title = $2;
            my $noexport = $title =~ s/\s*<!--\s*noexport\s*-->\s*$//i;
            if ($skip_level && $level > $skip_level) {
                next;
            }
            $skip_level = 0 if $skip_level && $level <= $skip_level;
            if ($noexport) {
                $skip_level = $level;
                next;
            }
            _markdown_heading($doc, $level, $title);
            next;
        }

        next if $skip_level;

        if ($line =~ /^\*\*([^*]+)\*\*$/ && !defined $doc->{meta}{subtitle}) {
            $doc->{meta}{subtitle} = $1;
            next;
        }
        if ($line =~ /^\*by\s+(.+)\*$/i) {
            $doc->{meta}{author} = $1;
            next;
        }
        if ($line =~ /^---\s+(.+?)\s+---$/) {
            my $content = $1;
            if ($content =~ /^(.+?)\s*\|\s*(.+)$/) {
                $doc->{meta}{version} = $1;
                $doc->{meta}{date} = $2;
            } else {
                $doc->{meta}{version} = $content;
            }
            next;
        }
        if ($line =~ /^\[\^(\S+)\]:\s*(.+)$/) {
            $doc->{footnote}{$1} = $2;
            next;
        }
        if ($line =~ /^\s*$/) {
            $flush->();
            next;
        }
        if ($line =~ /^\*\*\*$/) {
            $flush->();
            push @{$doc->{elements}}, { type => 'scene_break' };
            next;
        }

        push @paragraph, $line;
    }

    $flush->();
    return $doc;
}

sub _markdown_heading {
    my ($doc, $level, $title) = @_;

    if ($level == 1 && !defined $doc->{meta}{title}) {
        $doc->{meta}{title} = $title;
        return;
    }

    my $type = $level == 2 ? 'part' : $level == 3 ? 'chapter' : 'section';
    push @{$doc->{elements}}, { type => $type, title => $title };
    push @{$doc->{toc}}, { type => $type, title => $title };
}

sub _parse_org {
    my ($source) = @_;
    my $doc = _new_doc();
    my @lines = split /\n/, $source;
    my $skip_level = 0;
    my @paragraph;

    my $flush = sub {
        return if !@paragraph;
        push @{$doc->{elements}}, { type => 'paragraph', text => join("\n", @paragraph) };
        @paragraph = ();
    };

    for my $line (@lines) {
        if ($line =~ /^#\+([A-Z_]+):\s*(.*)$/i) {
            my $key = lc($1);
            $key =~ s/_/-/g;
            $doc->{meta}{$key} = $2;
            next;
        }
        if ($line =~ /^(\*+)\s+(.+)$/) {
            $flush->();
            my $level = length($1);
            my $title = $2;
            my $noexport = $title =~ s/\s*:noexport:\s*$//i;
            if ($skip_level && $level > $skip_level) {
                next;
            }
            $skip_level = 0 if $skip_level && $level <= $skip_level;
            if ($noexport) {
                $skip_level = $level;
                next;
            }
            my $type = $level == 1 ? 'part' : $level == 2 ? 'chapter' : 'section';
            push @{$doc->{elements}}, { type => $type, title => $title };
            push @{$doc->{toc}}, { type => $type, title => $title };
            next;
        }

        next if $skip_level;

        if ($line =~ /^\[fn:(\S+)\]\s+(.+)$/) {
            $doc->{footnote}{$1} = $2;
            next;
        }
        if ($line =~ /^-{5,}\s*$/) {
            $flush->();
            push @{$doc->{elements}}, { type => 'scene_break' };
            next;
        }
        if ($line =~ /^\s*$/) {
            $flush->();
            next;
        }

        push @paragraph, $line;
    }

    $flush->();
    return $doc;
}

sub _apply_config_metadata {
    my ($doc, $config, $metadata) = @_;

    for my $key (qw(title subtitle author date version wordcount address email website)) {
        my $value = $config->get($key);
        $doc->{meta}{$key} = $value if defined $value && $value ne '';
    }
    for my $key (qw(title subtitle author date version wordcount address email website)) {
        my $value = $metadata->{$key};
        $doc->{meta}{$key} = $value if defined $value && $value ne '';
    }
}

sub _render_typst {
    my ($doc, $config) = @_;

    my @out;
    push @out, '// Generated by First Folio manuscript';
    push @out, _typst_preamble($config, $doc);
    push @out, _title_page($config, $doc);
    push @out, _toc($config, $doc);
    push @out, _body($config, $doc);
    return join("\n", grep { defined $_ } @out) . "\n";
}

sub _typst_preamble {
    my ($config, $doc) = @_;

    my $page = _cfg($config, 'page', 'a4');
    my $margin = _cfg($config, 'margin', '20mm');
    my $font = _font_cfg($config, 'font', 'Libertinus Serif');
    my $font_size = _font_cfg($config, 'font-size', '12pt');
    my $mono_font = _cfg($config, 'mono-font', 'Libertinus Mono');
    my $title_font = _cfg($config, 'title-font', _heading_font($config));
    my $header = _cfg($config, 'page-header.format', '[author] / [title] / [page]');
    my $header_font = _cfg($config, 'page-header.font', _heading_font($config));
    my $header_size = _cfg($config, 'page-header.font-size', '10pt');
    my $header_edge = _cfg($config, 'page-header.distance-from-edge', $margin);
    my $header_pad = _cfg($config, 'page-header.content-padding-after', '10mm');
    my $line_spacing = _cfg($config, 'line-spacing', '1.5');
    my $para_indent = _cfg($config, 'paragraph-indent', '10mm');
    my $para_spacing = _cfg($config, 'paragraph-spacing', '0');
    my $typst_para_spacing = $para_spacing eq '0' ? '0pt' : $para_spacing;
    my $header_text = _fill_tokens($header, $doc);
    my $header_clause = '';
    if (_bool_cfg($config, 'page-header.enabled', 1)) {
        $header_clause = qq{, header-ascent: $header_edge, header: align(right)[#text(font: "$header_font", size: $header_size)[$header_text]]};
    }

    return join("\n",
        '// style: ' . _cfg($config, 'style', 'british'),
        "// margin: $margin",
        "// distance-from-edge: $header_edge",
        "// content-padding-after: $header_pad",
        "// page-header-format: $header",
        "// mono-font: $mono_font",
        "// title-font: $title_font",
        "// line-spacing: $line_spacing",
        "// paragraph-indent: $para_indent",
        "// paragraph-spacing: $para_spacing",
        qq{#set page(paper: "$page", margin: $margin, numbering: "1"$header_clause)},
        qq{#set text(font: "$font", size: $font_size)},
        qq{#set par(first-line-indent: $para_indent, spacing: $typst_para_spacing, leading: ${line_spacing}em)},
    );
}

sub _title_page {
    my ($config, $doc) = @_;

    return '' if !_bool_cfg($config, 'title-page.enabled', 1);
    my $title = _meta($doc, 'title');
    return '' if $title eq '';

    my $title_font = _cfg($config, 'title-font', _heading_font($config));
    my $title_size = _cfg($config, 'title-font-size', '20pt');
    my $title_weight = _cfg($config, 'title-font-weight', 'bold');
    my $subtitle = _meta($doc, 'subtitle');
    my $subtitle_font = _cfg($config, 'subtitle-font', _heading_font($config));
    my $subtitle_size = _cfg($config, 'subtitle-font-size', '14pt');
    my $subtitle_style = _cfg($config, 'subtitle-font-style', 'normal');
    my $author = _meta($doc, 'author');
    my $author_attr = _cfg($config, 'author-attribution', 'by');
    my $author_font = _cfg($config, 'author-font', _heading_font($config));
    my $date = _meta($doc, 'date');
    my $version = _meta($doc, 'version');
    my $wordcount = _meta($doc, 'wordcount');
    my $address = _meta($doc, 'address');
    my $email = _meta($doc, 'email');
    my $website = _meta($doc, 'website');
    my $header_pad = _cfg($config, 'page-header.content-padding-after', '10mm');

    my @lines;
    push @lines, '#set page(numbering: none)';
    push @lines, "#v($header_pad)";
    push @lines, '#align(center)[';
    if (_bool_cfg($config, 'title-page.include-title', 1)) {
        push @lines, qq{#text(font: "$title_font", size: $title_size, weight: "$title_weight")[@{[_esc($title)]}]};
    }
    if ($subtitle ne '' && _bool_cfg($config, 'title-page.include-subtitle', 1)) {
        push @lines, qq{#v(1em)\n#text(font: "$subtitle_font", size: $subtitle_size, style: "$subtitle_style")[@{[_esc($subtitle)]}]};
    }
    if ($author ne '' && _bool_cfg($config, 'title-page.include-author', 1)) {
        push @lines, qq{#v(2em)\n#text(font: "$author_font")[@{[_esc($author_attr)]} @{[_esc($author)]}]};
    }
    push @lines, ']';
    push @lines, '#v(1fr)';
    push @lines, _meta_line($config, 'date', $date)
        if _bool_cfg($config, 'title-page.include-date', 1);
    push @lines, _meta_line($config, 'version', $version)
        if _bool_cfg($config, 'title-page.include-version', 1);
    push @lines, _meta_line($config, 'wordcount', "word count: $wordcount")
        if $wordcount ne '' && _bool_cfg($config, 'title-page.include-wordcount', 1);
    push @lines, _meta_line($config, 'contact', $address)
        if _bool_cfg($config, 'title-page.include-address', 1);
    push @lines, _meta_line($config, 'contact', $email)
        if _bool_cfg($config, 'title-page.include-email', 1);
    push @lines, _meta_line($config, 'contact', $website)
        if _bool_cfg($config, 'title-page.include-website', 1);
    push @lines, '#pagebreak()';
    push @lines, '#set page(numbering: "1")';
    return join("\n", grep { defined $_ && $_ ne '' } @lines);
}

sub _meta_line {
    my ($config, $prefix, $text) = @_;
    return '' if !defined $text || $text eq '';

    my $font = _cfg($config, "$prefix-font", _heading_font($config));
    my $size = _cfg($config, "$prefix-font-size", '10pt');
    my $weight = _cfg($config, "$prefix-font-weight", 'regular');
    return qq{#text(font: "$font", size: $size, weight: "$weight")[@{[_esc($text)]}] \\};
}

sub _toc {
    my ($config, $doc) = @_;

    return '' if !_bool_cfg($config, 'toc.enabled', 1);
    my $title = _cfg($config, 'toc.title', 'Contents');
    my $font = _cfg($config, 'toc.font', _heading_font($config));
    my $font_size = _cfg($config, 'toc.font-size', '11pt');
    my $heading_font = _cfg($config, 'toc.heading-font', _heading_font($config));
    my $heading_size = _cfg($config, 'toc.heading-font-size', '16pt');
    my $include_parts = _bool_cfg($config, 'toc.include-parts', 1);
    my $include_chapters = _bool_cfg($config, 'toc.include-chapters', 1);
    my $include_sections = _bool_cfg($config, 'toc.include-sections', 0);
    my @lines;
    push @lines, qq{// TOC Font: $font};
    push @lines, qq{// TOC Heading: $heading_font};
    push @lines, qq{#text(font: "$heading_font", size: $heading_size)[@{[_esc($title)]}]};
    for my $entry (@{$doc->{toc}}) {
        next if $entry->{type} eq 'part' && !$include_parts;
        next if $entry->{type} eq 'chapter' && !$include_chapters;
        next if $entry->{type} eq 'section' && !$include_sections;
        push @lines, qq{#text(font: "$font", size: $font_size)[@{[_esc($entry->{title})]}]};
    }
    push @lines, '#pagebreak()';
    return join("\n", @lines);
}

sub _body {
    my ($config, $doc) = @_;

    my @lines;
    my $heading_font = _heading_font($config);
    my $heading_size = _heading_size($config);
    for my $el (@{$doc->{elements}}) {
        if ($el->{type} eq 'part') {
            push @lines, '#pagebreak()';
            push @lines, '// vertical-align: center';
            push @lines, qq{#align(center + horizon)[#text(font: "$heading_font", size: $heading_size)[@{[_esc(_case($config, 'part', $el->{title}))]}]]};
        } elsif ($el->{type} eq 'chapter') {
            push @lines, '#pagebreak()';
            push @lines, '// position: one-third';
            push @lines, qq{#v(33%)\n#align(center)[#text(font: "$heading_font", size: $heading_size)[@{[_esc($el->{title})]}]]};
        } elsif ($el->{type} eq 'section') {
            push @lines, qq{#text(font: "$heading_font", size: $heading_size)[@{[_esc($el->{title})]}]};
        } elsif ($el->{type} eq 'scene_break') {
            push @lines, '#align(center)[\\* \\* \\*] // scene-break';
        } elsif ($el->{type} eq 'code_start') {
            my $mono = _cfg($config, 'mono-font', 'Libertinus Mono');
            push @lines, qq{// mono-font: $mono};
            push @lines, qq{#text(font: "$mono")[#raw(block: true, "};
        } elsif ($el->{type} eq 'code_end') {
            push @lines, '")]';
        } elsif ($el->{type} eq 'code') {
            push @lines, _raw_esc($el->{text});
        } elsif ($el->{type} eq 'paragraph') {
            push @lines, _inline($el->{text}, $doc, $config);
            push @lines, '';
        }
    }
    return join("\n", @lines);
}

sub _cfg {
    my ($config, $key, $default) = @_;
    my $val = $config->get("folio.manuscript.$key");
    return $val if defined $val;

    if ($key =~ /^(font|font-size|font-weight|font-stretch|page|margin)$/) {
        my $root = $config->get("folio.$key");
        return $root if defined $root;
    }
    return $default;
}

sub _font_cfg {
    my ($config, $key, $default) = @_;
    return _cfg($config, $key, $default);
}

sub _heading_font {
    my ($config) = @_;
    return _cfg($config, 'heading-font', $config->get('folio.heading-font') // _cfg($config, 'font', 'Libertinus Serif'));
}

sub _heading_size {
    my ($config) = @_;
    return _cfg($config, 'heading-font-size', $config->get('folio.heading-font-size') // _cfg($config, 'font-size', '12pt'));
}

sub _bool_cfg {
    my ($config, $key, $default) = @_;
    my $val = $config->get("folio.manuscript.$key");
    return $default if !defined $val;
    return $val ? 1 : 0;
}

sub _meta {
    my ($doc, $key) = @_;
    return $doc->{meta}{$key} // '';
}

sub _fill_tokens {
    my ($template, $doc) = @_;
    my %values = (
        author => _meta($doc, 'author'),
        title  => _meta($doc, 'title'),
        page   => '#context counter(page).display()',
    );
    $template =~ s/\[(author|title|page)\]/$values{$1}/g;
    return _esc($template);
}

sub _case {
    my ($config, $kind, $text) = @_;
    my $case = _cfg($config, "$kind.case-transform", 'as-written');
    return uc($text) if $case eq 'upper';
    return $text;
}

sub _inline {
    my ($text, $doc, $config) = @_;
    my $mono = _cfg($config, 'mono-font', 'Libertinus Mono');
    my @footnotes;
    $text =~ s/\[\^(\S+?)\]/_stash_footnote(\@footnotes, _footnote($doc, $1))/ge;
    $text =~ s/\[fn:(\S+?)\]/_stash_footnote(\@footnotes, _footnote($doc, $1))/ge;
    my @raw;
    $text =~ s/`([^`]+)`/_stash_raw(\@raw, _inline_raw($mono, $1))/ge;
    $text =~ s/~([^~]+)~/_stash_raw(\@raw, _inline_raw($mono, $1))/ge;
    $text = _esc($text);
    $text =~ s{\*\*([^*\n]+)\*\*}{*$1*}g;
    $text =~ s{(?<!\w)\*([^*\n]+)\*(?!\w)}{_$1_}g;
    $text =~ s{(?<!\w)/([^/\n]+)/(?!\w)}{_$1_}g;
    for my $i (0 .. $#raw) {
        my $token = "__FOLIO_RAW_${i}__";
        $text =~ s/$token/$raw[$i]/g;
    }
    for my $i (0 .. $#footnotes) {
        my $token = "__FOLIO_FOOTNOTE_${i}__";
        $text =~ s/$token/$footnotes[$i]/g;
    }
    return $text;
}

sub _stash_raw {
    my ($raw, $text) = @_;
    push @$raw, $text;
    my $index = $#$raw;
    return "__FOLIO_RAW_${index}__";
}

sub _inline_raw {
    my ($font, $text) = @_;
    return qq{#text(font: "$font")[#raw("@{[_raw_esc($text)]}")]};
}

sub _stash_footnote {
    my ($footnotes, $text) = @_;
    push @$footnotes, $text;
    my $index = $#$footnotes;
    return "__FOLIO_FOOTNOTE_${index}__";
}

sub _footnote {
    my ($doc, $name) = @_;
    my $text = $doc->{footnote}{$name} // $name;
    return '#footnote[' . _esc($text) . ']';
}

sub _write_text {
    my ($path, $text) = @_;
    open(my $fh, '>:encoding(UTF-8)', $path)
        or die "Error: cannot write to $path: $!\n";
    print $fh $text;
    close $fh;
}

sub _esc {
    my ($text) = @_;
    return '' if !defined $text;
    $text =~ s/\\/\\\\/g;
    $text =~ s/\[/\\[/g;
    $text =~ s/\]/\\]/g;
    $text =~ s/\$/\\\$/g;
    $text =~ s/@/\\@/g;
    return $text;
}

sub _raw_esc {
    my ($text) = @_;
    $text =~ s/\\/\\\\/g;
    $text =~ s/"/\\"/g;
    return $text;
}

1;
