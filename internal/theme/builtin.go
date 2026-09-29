package theme

// builtins ship with acta. The hex comes from the Ghostty files on
// terminalcolors.com, so they match what users see in their terminal.
var builtins = map[string]Theme{
	"terminal": {},
	"tokyo-night": {BG: "#1a1b26", FG: "#c0caf5", SelectionBG: "#283457", SelectionFG: "#c0caf5",
		ANSI: [16]string{"#15161e", "#f7768e", "#9ece6a", "#e0af68", "#7aa2f7", "#bb9af7", "#7dcfff", "#a9b1d6",
			"#414868", "#f7768e", "#9ece6a", "#e0af68", "#7aa2f7", "#bb9af7", "#7dcfff", "#c0caf5"}},
	"tokyo-night-day": {BG: "#e1e2e7", FG: "#3760bf", SelectionBG: "#b7c1e3", SelectionFG: "#3760bf",
		ANSI: [16]string{"#b4b5b9", "#f52a65", "#587539", "#8c6c3e", "#2e7de9", "#9854f1", "#007197", "#6172b0",
			"#a1a6c5", "#f52a65", "#587539", "#8c6c3e", "#2e7de9", "#9854f1", "#007197", "#3760bf"}},
	"catppuccin-mocha": {BG: "#1e1e2e", FG: "#cdd6f4", SelectionBG: "#353748", SelectionFG: "#cdd6f4",
		ANSI: [16]string{"#45475a", "#f38ba8", "#a6e3a1", "#f9e2af", "#89b4fa", "#f5c2e7", "#94e2d5", "#a6adc8",
			"#585b70", "#f37799", "#89d88b", "#ebd391", "#74a8fc", "#f2aede", "#6bd7ca", "#bac2de"}},
	"catppuccin-latte": {BG: "#eff1f5", FG: "#4c4f69", SelectionBG: "#d8dae1", SelectionFG: "#4c4f69",
		ANSI: [16]string{"#5c5f77", "#d20f39", "#40a02b", "#df8e1d", "#1e66f5", "#ea76cb", "#179299", "#acb0be",
			"#6c6f85", "#de293e", "#49af3d", "#eea02d", "#456eff", "#fe85d8", "#2d9fa8", "#bcc0cc"}},
	"gruvbox-dark": {BG: "#282828", FG: "#ebdbb2", SelectionBG: "#ebdbb2", SelectionFG: "#282828",
		ANSI: [16]string{"#282828", "#cc241d", "#98971a", "#d79921", "#458588", "#b16286", "#689d6a", "#a89984",
			"#928374", "#fb4934", "#b8bb26", "#fabd2f", "#83a598", "#d3869b", "#8ec07c", "#ebdbb2"}},
	"dracula": {BG: "#282a36", FG: "#f8f8f2", SelectionBG: "#44475a", SelectionFG: "#f8f8f2",
		ANSI: [16]string{"#21222c", "#ff5555", "#50fa7b", "#f1fa8c", "#bd93f9", "#ff79c6", "#8be9fd", "#f8f8f2",
			"#6272a4", "#ff6e6e", "#69ff94", "#ffffa5", "#d6acff", "#ff92df", "#a4ffff", "#ffffff"}},
}
