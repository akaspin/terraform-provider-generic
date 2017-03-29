resource "generic_shell" "test" {
  trigger {
    a = "1"
  }
  inline = [
    "#!/usr/bin/env bash",
    "echo a > test"
  ]
}

resource "generic_shell" "test1" {
  inline = [
    "rm test"
  ]
  phase = "destroy"
}