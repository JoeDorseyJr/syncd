package defaults

import "testing"

func TestCompareValue_IntMatch(t *testing.T) {
	if !CompareValue("48", "int", 48) {
		t.Fatal("expected match")
	}
}

func TestCompareValue_IntMismatch(t *testing.T) {
	if CompareValue("49", "int", 48) {
		t.Fatal("expected no match")
	}
}

func TestCompareValue_FloatMatch(t *testing.T) {
	if !CompareValue("0.5", "float", 0.5) {
		t.Fatal("expected match")
	}
}

func TestCompareValue_BoolTrueMatch(t *testing.T) {
	if !CompareValue("1", "bool", true) {
		t.Fatal("expected match")
	}
}

func TestCompareValue_BoolFalseMatch(t *testing.T) {
	if !CompareValue("0", "bool", false) {
		t.Fatal("expected match")
	}
}

func TestCompareValue_BoolMismatch(t *testing.T) {
	if CompareValue("0", "bool", true) {
		t.Fatal("expected no match")
	}
}

func TestCompareValue_StringMatch(t *testing.T) {
	if !CompareValue("hello", "string", "hello") {
		t.Fatal("expected match")
	}
}

func TestCompareValue_StringMismatch(t *testing.T) {
	if CompareValue("hello", "string", "world") {
		t.Fatal("expected no match")
	}
}
