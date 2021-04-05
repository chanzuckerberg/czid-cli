#!/bin/bash

mkdir completions

$1 completion bash > completions/bash
$1 completion zsh > completions/zsh
$1 completion fish > completions/fish
$1 completion powershell > completions/powershell
