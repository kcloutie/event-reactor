#!/bin/env pwsh

$env:dockerCmd = "podman"
# $env:dockerCmd = "docker"

$ProgressPreference = "SilentlyContinue"
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version 1.0
Trap {
  CleanUp
  $_; Write-Output "AN ERROR HAS OCCURRED!!"
  Exit 1 
}

function New-TemporaryDirectory {
  $parent = [System.IO.Path]::GetTempPath()
  $name = "EVENT_REACTOR_$([System.IO.Path]::GetRandomFileName())"
  New-Item -ItemType Directory -Path (Join-Path $parent $name)
}



Set-Location -Path (Split-Path -Parent -Path ((Resolve-Path -LiteralPath $MyInvocation.MyCommand.Definition).Path))

if (-not (Test-Path env:KIND_CLUSTER_NAME)) {
  $env:KIND_CLUSTER_NAME = "kind"
}
$env:KUBECONFIG = "${env:USERPROFILE}\.kube\config.${env:KIND_CLUSTER_NAME}"
$env:TARGET="kubernetes"
$env:DOMAIN_NAME="paac-127-0-0-1.nip.io"

if (-not (Get-Command kind -ErrorAction SilentlyContinue)) {
  Write-Output "Install kind. https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
  exit 1
}
$env:kind = Get-Command kind -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source

if (-not (Get-Command ko -ErrorAction SilentlyContinue)) {
  Write-Output "Install ko. https://ko.build/install/"
  exit 1
}
$env:ko = Get-Command ko -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source

$env:TMPD = New-TemporaryDirectory | Select-Object -ExpandProperty FullName

$env:REG_PORT='5000'
$env:REG_NAME='kind-registry'

function CleanUp {
  if ($TMPD) {
    Remove-Item -Path $TMPD -Recurse -Force
  }  
}

function Start-Registry {
  $running = &$env:dockerCmd inspect --format='{{.State.Running}}' $env:REG_NAME 2>$null
  if ($LASTEXITCODE -ne 0) {
      $running = $false
  }

  if (!$running) {
    &$env:dockerCmd rm -f kind-registry 2>$null
    &$env:dockerCmd run -d --restart=always -p "127.0.0.1:$($env:REG_PORT):5000" -e REGISTRY_HTTP_SECRET=secret --name $env:REG_NAME  registry:2  
  }
}

function ReinstallKind {
  &$env:kind delete cluster --name $env:KIND_CLUSTER_NAME 2>$null
  &$env:kind create cluster --name $env:KIND_CLUSTER_NAME --config  $env:TMPD/kconfig.yaml
}