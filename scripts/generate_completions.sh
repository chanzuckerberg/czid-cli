#!/bin/bash

$1 completion bash > ./bash_completions
$1 completion zsh > ./zsh_completions
$1 completion fish > ./fish_completions
$1 completion powershell > ./powershell_completions
