package main

import (
  "unicode"
)

func getAnsiFor(c string) string {
  var col string
  switch (unicode.ToLower(rune(c[2]))) {
  case 'a':
    col = "92"
  case 'b':
    col = "36"
  case 'c':
    col = "91"
  case 'd':
    col = "95"
  case 'e':
    col = "93"
  case 'f':
    col = "97"
  case '0':
    col = "90"
  case '1':
    col = "34"
  case '2':
    col = "32"
  case '3':
    col = "36"
  case '4':
    col = "91"
  case '5':
    col = "35"
  case '6':
    col = "33"
  case '7':
    col = "37"
  case '8':
    col = "90"
  case '9':
    col = "94"
  case 'l':
    col = "1"
  case 'm':
    col = "2"
  case 'n':
    col = "4"
  case 'k':
    col = "7"
  case 'r':
    col = "0"
  }
  colorStr := "\x1B\x5B" + col + "m"
  switch (unicode.ToLower(rune(c[2]))) {
  case 'a', 'b', 'c', 'd', 'e', 'f', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
    colorStr = "\x1B\x5B0m" + colorStr
  }
  return colorStr
}
