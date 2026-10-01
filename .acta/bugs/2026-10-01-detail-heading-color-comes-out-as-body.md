---
id: BUG-0027
hash: qknr0rt
priority: medium
---
# Detail markdown heading paints in the body color, not the heading color

## Symptom
The heading in the detail pane markdown is painted in the theme foreground
(#c0caf5 on tokyo-night) instead of the heading yellow the style config asks
for. Bold still shows red and inline code still shows green, so only the
heading is wrong.

## Repro
`markdownStyle(hexTheme, true).Heading.Color` is `#e0af68`, but the rendered
output of `newRenderer(hexTheme, true)("# H", 80)` holds no `38;2;224;175;104`
at all and paints the `H` with `38;2;192;202;245;1m`, which is the foreground.

## Cause
`markdownStyle` sets `cfg.Text.Color = body`, and glamour paints the plain
text inside a heading with `Styles.Text` as the child style. In
`cascadeStylePrimitive` the child wins when it has a color, so the body color
overwrites the yellow the heading block carried. `Emphasis` and `Code` use
`StyleOverrideRender`, which is why those two still show their own colors.

## Where
`internal/tui/markdown.go`, the `cfg.Text.Color = body` line.
