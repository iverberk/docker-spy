package main

func TruncateString(str string, maxLength int) string {
  runes := []rune(str)
  if len(runes) > maxLength {
    return string(runes[:maxLength])
  }

  return str
}