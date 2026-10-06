package stub

import (
	"strings"
	"testing"
)

var twoSum = Function{
	Name:    "two_sum",
	Input:   []string{"n: int", "target: int", "nums: int[n]"},
	Params:  []string{"nums", "target"},
	Returns: "int[]",
}

func TestGenerate_EveryLanguage(t *testing.T) {
	want := map[string]string{
		"c":          "int* two_sum(int* nums, int nums_size, int target, int* return_size)",
		"cpp":        "vector<int> twoSum(vector<int>& nums, int target)",
		"java":       "public int[] twoSum(int[] nums, int target)",
		"python":     "def two_sum(nums: list[int], target: int) -> list[int]:",
		"go":         "func twoSum(nums []int, target int) []int",
		"rust":       "pub fn two_sum(nums: Vec<i32>, target: i32) -> Vec<i32>",
		"javascript": "function twoSum(nums, target)",
	}
	for lang, signature := range want {
		code, err := Generate(lang, twoSum)
		if err != nil {
			t.Fatalf("%s: %v", lang, err)
		}
		if !strings.Contains(code, signature) {
			t.Errorf("%s: no %q in\n%s", lang, signature, code)
		}
		if !strings.Contains(code, "No need to change it") {
			t.Errorf("%s: no banner between the function and main", lang)
		}
	}
}

func TestCheck(t *testing.T) {
	bad := []Function{
		{Name: "TwoSum", Input: []string{"n: int"}, Params: []string{"n"}, Returns: "int"},
		{Name: "f", Input: []string{"nums: int[n]"}, Params: []string{"nums"}, Returns: "int"},
		{Name: "f", Input: []string{"n: int", "n: int"}, Params: []string{"n"}, Returns: "int"},
		{Name: "f", Input: []string{"n: int"}, Params: []string{"m"}, Returns: "int"},
		{Name: "f", Input: []string{"s: line[2]"}, Params: []string{"s"}, Returns: "int"},
		{Name: "f", Input: []string{"n: int", "g: int[n][]"}, Params: []string{"g"}, Returns: "int"},
		{Name: "f", Input: []string{"n: int"}, Params: []string{"n"}, Returns: "int[n]"},
		{Name: "f", Input: []string{"n: int"}, Params: []string{"n"}, Returns: "word"},
		{Name: "f", Input: []string{"n: int"}, Params: []string{"n"}, Returns: "int", Output: "lines"},
		{Name: "f", Input: []string{"n: int"}, Params: []string{"n"}, Returns: "int[]", Output: "count"},
		{Name: "f", Input: []string{"b: bool"}, Params: []string{"b"}, Returns: "int"},
	}
	for _, f := range bad {
		if err := f.Check(); err == nil {
			t.Errorf("%+v passed the check", f)
		}
	}
	if err := twoSum.Check(); err != nil {
		t.Errorf("two sum: %v", err)
	}
}

func TestCamel(t *testing.T) {
	for in, want := range map[string]string{"two_sum": "twoSum", "f": "f", "total_n_queens": "totalNQueens"} {
		if got := camel(in); got != want {
			t.Errorf("camel(%q) = %q, want %q", in, got, want)
		}
	}
}
