#!/bin/bash

$1 completion bash > bash_completion
$1 completion zsh > zsh_completion
$1 completion fish > fish_completion
$1 completion powershell > powershell_completion
