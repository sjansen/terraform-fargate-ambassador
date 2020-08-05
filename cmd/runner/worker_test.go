package main

import "testing"

func TestRot13(t *testing.T) {
	for s, expected := range map[string]string{
		"Gur jbeyq vf n inzcver.":          "The world is a vampire.",
		"⚠️Qb abg gnhag Unccl Sha Onyy.⚠️": "⚠️Do not taunt Happy Fun Ball.⚠️",
	} {
		t.Run(s, func(t *testing.T) {
			s, expected := s, expected
			actual := rot13(s)
			if expected != actual {
				t.Fatalf("expected=%v actual=%v", expected, actual)
			}
		})
	}
}
