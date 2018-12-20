workflow "Build" {
  on = "push"
  resolves = [
    "Crossbinary",
    "Go Build",
  ]
}

action "Clean" {
  uses = "./.github/actions/golang"
  args = "clean"
}

action "Go Test" {
  needs = ["Clean"]
  uses = "./.github/actions/golang"
  args = "test"
}

action "Go Build" {
  needs = ["Go Test"]
  uses = "./.github/actions/golang"
  args = "build"
}

action "Crossbinary" {
  needs = ["Go Test"]
  uses = "./.github/actions/golang"
  args = "build-crossbinary"
}