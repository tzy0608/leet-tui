package editor

import "fmt"

// CodeTemplate returns a boilerplate template for the given language.
func CodeTemplate(lang, functionSig string) string {
	if functionSig != "" {
		return functionSig
	}

	switch lang {
	case "go", "golang":
		return "package main\n\nfunc solution() {\n\t// TODO: implement\n}\n"
	case "python", "python3":
		return "class Solution:\n    def solve(self):\n        # TODO: implement\n        pass\n"
	case "cpp", "c++":
		return "class Solution {\npublic:\n    // TODO: implement\n};\n"
	case "java":
		return "class Solution {\n    // TODO: implement\n}\n"
	default:
		return fmt.Sprintf("// %s solution\n// TODO: implement\n", lang)
	}
}
