#!/bin/bash

mkdir $1/completions

$2 completion bash > $1/completions/bash
$2 completion zsh > $1/completions/zsh
$2 completion fish > $1/completions/fish
$2 completion powershell > $1/completions/powershell
