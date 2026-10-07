[CmdletBinding()]
param(
  [ValidateSet('contract', 'release')]
  [string]$Gate = 'contract'
)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot

function Require-File([string]$RelativePath) {
  $path = Join-Path $root $RelativePath
  if (-not (Test-Path -LiteralPath $path)) { throw "Required release artifact is missing: $RelativePath" }
  return $path
}

if ($Gate -eq 'contract') {
  $openApiPath = Require-File 'docs/openapi.v1.json'
  $openApi = Get-Content -Raw -LiteralPath $openApiPath | ConvertFrom-Json
  if ($openApi.openapi -notmatch '^3\.1\.0$') { throw 'OpenAPI document must declare version 3.1.0.' }
  foreach ($path in '/api/projects', '/api/projects/{id}/events', '/api/work-items/{key}', '/api/api-requests/{id}/runs') {
    if (-not $openApi.paths.PSObject.Properties[$path]) { throw "OpenAPI is missing implemented path $path" }
  }
  Write-Host "OpenAPI $($openApi.info.version) is structurally valid."
  exit 0
}

foreach ($artifact in 'docs/RUNBOOK.md', 'docs/RELEASE-CHECKLIST.md', 'docs/ARCHITECTURE-ERD.md', 'docs/MIGRATION-REHEARSAL.md', 'docs/RELEASE-EVIDENCE.md', 'docs/RELEASE-EVIDENCE-2026-10-03.md') { [void](Require-File $artifact) }
$trackedSQLite = & git -C $root ls-files --error-unmatch backend/flowak.db 2>$null
if ($LASTEXITCODE -eq 0 -and $trackedSQLite) { throw 'Tracked SQLite runtime artifact found; PostgreSQL must remain the only runtime database.' }
$global:LASTEXITCODE = 0
$localSQLite = Get-ChildItem -Path (Join-Path $root 'backend') -Recurse -File -Filter 'flowak.db' -ErrorAction SilentlyContinue
if ($localSQLite) { Write-Warning 'Ignored local SQLite artifact found. It is not part of this release branch and must not be used as a runtime database.' }
Write-Host 'Release documents, PostgreSQL-only runtime assertion, and evidence template are present.'
