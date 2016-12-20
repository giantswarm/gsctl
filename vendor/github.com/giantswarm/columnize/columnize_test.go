package columnize

import (
	"fmt"
	"reflect"
	"testing"
)

func TestListOfStringsInput(t *testing.T) {
	input := []string{
		"Column A | Column B | Column C",
		"x | y | z",
	}

	config := DefaultConfig()
	output := Format(input, config)

	expected := "Column A  Column B  Column C\n"
	expected += "x         y         z"

	if output != expected {
		t.Fatalf("\nexpected:\n%s\n\ngot:\n%s", expected, output)
	}
}

func TestEmptyLinesOutput(t *testing.T) {
	input := []string{
		"Column A | Column B | Column C",
		"",
		"x | y | z",
	}

	config := DefaultConfig()
	output := Format(input, config)

	expected := "Column A  Column B  Column C\n"
	expected += "\n"
	expected += "x         y         z"

	if output != expected {
		t.Fatalf("\nexpected:\n%s\n\ngot:\n%s", expected, output)
	}
}

func TestLeadingSpacePreserved(t *testing.T) {
	input := []string{
		"| Column B | Column C",
		"x | y | z",
	}

	config := DefaultConfig()
	output := Format(input, config)

	expected := "   Column B  Column C\n"
	expected += "x  y         z"

	if output != expected {
		t.Fatalf("\nexpected:\n%s\n\ngot:\n%s", expected, output)
	}
}

func TestColumnWidthCalculator(t *testing.T) {
	input := []string{
		"Column A | Column B | Column C",
		"Longer than A | Longer than B | Longer than C",
		"short | short | short",
	}

	config := DefaultConfig()
	output := Format(input, config)

	expected := "Column A       Column B       Column C\n"
	expected += "Longer than A  Longer than B  Longer than C\n"
	expected += "short          short          short"

	if output != expected {
		t.Fatalf("\nexpected:\n%s\n\ngot:\n%s", expected, output)
	}
}

func TestColumnWidthCalculatorNonASCII(t *testing.T) {
	input := []string{
		"Column A | Column B | Column C",
		"⌘⌘⌘⌘⌘⌘⌘⌘ | Longer than B | Longer than C",
		"short | short | short",
	}

	config := DefaultConfig()
	output := Format(input, config)

	expected := "Column A  Column B       Column C\n"
	expected += "⌘⌘⌘⌘⌘⌘⌘⌘  Longer than B  Longer than C\n"
	expected += "short     short          short"

	if output != expected {
		printableProof := fmt.Sprintf("\nGot:      %+q", output)
		printableProof += fmt.Sprintf("\nExpected: %+q", expected)
		t.Fatalf("\n%s", printableProof)
	}
}

func TestVariedInputSpacing(t *testing.T) {
	input := []string{
		"Column A       |Column B|    Column C",
		"x|y|          z",
	}

	config := DefaultConfig()
	output := Format(input, config)

	expected := "Column A  Column B  Column C\n"
	expected += "x         y         z"

	if output != expected {
		t.Fatalf("\nexpected:\n%s\n\ngot:\n%s", expected, output)
	}
}

func TestUnmatchedColumnCounts(t *testing.T) {
	input := []string{
		"Column A | Column B | Column C",
		"Value A | Value B",
		"Value A | Value B | Value C | Value D",
	}

	config := DefaultConfig()
	output := Format(input, config)

	expected := "Column A  Column B  Column C\n"
	expected += "Value A   Value B\n"
	expected += "Value A   Value B   Value C   Value D"

	if output != expected {
		t.Fatalf("\nexpected:\n%s\n\ngot:\n%s", expected, output)
	}
}

func TestAlternateDelimiter(t *testing.T) {
	input := []string{
		"Column | A % Column | B % Column | C",
		"Value A % Value B % Value C",
	}

	config := DefaultConfig()
	config.Delim = "%"
	output := Format(input, config)

	expected := "Column | A  Column | B  Column | C\n"
	expected += "Value A     Value B     Value C"

	if output != expected {
		t.Fatalf("\nexpected:\n%s\n\ngot:\n%s", expected, output)
	}
}

func TestAlternateSpacingString(t *testing.T) {
	input := []string{
		"Column A | Column B | Column C",
		"x | y | z",
	}

	config := DefaultConfig()
	config.Glue = "    "
	output := Format(input, config)

	expected := "Column A    Column B    Column C\n"
	expected += "x           y           z"

	if output != expected {
		t.Fatalf("\nexpected:\n%s\n\ngot:\n%s", expected, output)
	}
}

func TestSimpleFormat(t *testing.T) {
	input := []string{
		"Column A | Column B | Column C",
		"x | y | z",
	}

	output := SimpleFormat(input)

	expected := "Column A  Column B  Column C\n"
	expected += "x         y         z"

	if output != expected {
		t.Fatalf("\nexpected:\n%s\n\ngot:\n%s", expected, output)
	}
}

func TestAlternatePrefixString(t *testing.T) {
	input := []string{
		"Column A | Column B | Column C",
		"x | y | z",
	}

	config := DefaultConfig()
	config.Prefix = "  "
	output := Format(input, config)

	expected := "  Column A  Column B  Column C\n"
	expected += "  x         y         z"

	if output != expected {
		t.Fatalf("\nexpected:\n%s\n\ngot:\n%s", expected, output)
	}
}

func TestEmptyFieldReplacement(t *testing.T) {
	input := []string{
		"Column A | Column B | Column C",
		"x | | z",
	}

	config := DefaultConfig()
	config.Empty = "<none>"
	output := Format(input, config)

	expected := "Column A  Column B  Column C\n"
	expected += "x         <none>    z"

	if output != expected {
		t.Fatalf("\nexpected:\n%s\n\ngot:\n%s", expected, output)
	}
}

func TestEmptyConfigValues(t *testing.T) {
	input := []string{
		"Column A | Column B | Column C",
		"x | y | z",
	}

	config := Config{}
	output := Format(input, &config)

	expected := "Column A  Column B  Column C\n"
	expected += "x         y         z"

	if output != expected {
		t.Fatalf("\nexpected:\n%s\n\ngot:\n%s", expected, output)
	}
}

func TestMergeConfig(t *testing.T) {
	conf1 := &Config{Delim: "a", Glue: "a", Prefix: "a", Empty: "a"}
	conf2 := &Config{Delim: "b", Glue: "b", Prefix: "b", Empty: "b"}
	conf3 := &Config{Delim: "c", Prefix: "c"}

	m := MergeConfig(conf1, conf2)
	if m.Delim != "b" || m.Glue != "b" || m.Prefix != "b" || m.Empty != "b" {
		t.Fatalf("bad: %#v", m)
	}

	m = MergeConfig(conf1, conf3)
	if m.Delim != "c" || m.Glue != "a" || m.Prefix != "c" || m.Empty != "a" {
		t.Fatalf("bad: %#v", m)
	}

	m = MergeConfig(conf1, nil)
	if m.Delim != "a" || m.Glue != "a" || m.Prefix != "a" || m.Empty != "a" {
		t.Fatalf("bad: %#v", m)
	}

	m = MergeConfig(conf1, &Config{})
	if m.Delim != "a" || m.Glue != "a" || m.Prefix != "a" || m.Empty != "a" {
		t.Fatalf("bad: %#v", m)
	}
}

func TestGetWidthsFromLines01(t *testing.T) {
	config := DefaultConfig()
	input := []string{"first|line", "second|line", "third     |line"}
	expected := []int{6, 4}
	output := getWidthsFromLines(config, input)
	if !reflect.DeepEqual(output, expected) {
		printableProof := fmt.Sprintf("\nGot:      %s", output)
		printableProof += fmt.Sprintf("\nExpected: %s", expected)
		t.Fatalf("\n%s", printableProof)
	}
}

func TestGetWidthsFromLines02(t *testing.T) {
	config := DefaultConfig()
	input := []string{"\x1b[32mfirst\x1b[0m|line", "second|line"}
	expected := []int{6, 4}
	output := getWidthsFromLines(config, input)
	if !reflect.DeepEqual(output, expected) {
		printableProof := fmt.Sprintf("\nGot:      %s", output)
		printableProof += fmt.Sprintf("\nExpected: %s", expected)
		t.Fatalf("\n%s", printableProof)
	}
}

// testing width calculation for non-ASCII cahracters
func TestGetWidthsFromLines03(t *testing.T) {
	config := DefaultConfig()
	input := []string{"A|B", "⌘|⌘"}
	expected := []int{1, 1}
	output := getWidthsFromLines(config, input)
	if !reflect.DeepEqual(output, expected) {
		printableProof := fmt.Sprintf("\nGot:      %s", output)
		printableProof += fmt.Sprintf("\nExpected: %s", expected)
		t.Fatalf("\n%s", printableProof)
	}
}

// testing width calculation for strings with UTF-8 characters and color codes
func TestGetWidthsFromLines04(t *testing.T) {
	config := DefaultConfig()
	input := []string{"\x1b[32m⌘\x1b[0m|⌘", "A|B"}
	expected := []int{1, 1}
	output := getWidthsFromLines(config, input)
	if !reflect.DeepEqual(output, expected) {
		printableProof := fmt.Sprintf("\nGot:      %s", output)
		printableProof += fmt.Sprintf("\nExpected: %s", expected)
		t.Fatalf("\n%s", printableProof)
	}
}

func TestGetStringFormat01(t *testing.T) {
	config := DefaultConfig()
	widths := []int{13, 13, 3}
	elems := getElementsFromLine(config, "first element | second element | end")
	expected := "%-13s  %-13s  %s\n"
	output := config.getStringFormat(widths, elems)
	if output != expected {
		t.Fatalf("\nGot:      %s\nExpected: %s", output, expected)
	}
}
func TestGetStringFormat02(t *testing.T) {
	config := DefaultConfig()
	widths := []int{13, 13, 15, 3}
	elems := getElementsFromLine(config, "first element | second item ⌘ | third element \x1b[32m⌘\x1b[0m | end")
	output := config.getStringFormat(widths, elems)
	expected := "%-13s  %-13s  %-24s  %s\n"
	if output != expected {
		t.Fatalf("\nGot:      %q\nExpected: %q", output, expected)
	}
}

func TestDontCountColorCodes(t *testing.T) {
	input := []string{
		"\x1b[31;1mColumn A\x1b[0m | \x1b[32mColumn B\x1b[0m | \x1b[34mColumn C\x1b[0m",
		"Longer than A | Longer than B | Longer than C",
	}

	config := DefaultConfig()
	output := Format(input, config)

	expected := "\x1b[31;1mColumn A\x1b[0m       \x1b[32mColumn B\x1b[0m       \x1b[34mColumn C\x1b[0m\n"
	expected += "Longer than A  Longer than B  Longer than C"

	if output != expected {
		printableProof := fmt.Sprintf("\nGot:      %+q", output)
		printableProof += fmt.Sprintf("\nExpected: %+q", expected)
		t.Fatalf("\n%s", printableProof)
	}
}
