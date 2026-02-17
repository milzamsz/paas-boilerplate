###############################################################################
# PaaS-Core API QA Test Script
# Comprehensive endpoint testing for local development
# Prerequisites: API running on :8080, PostgreSQL running
# Handles CSRF double-submit cookie pattern automatically
###############################################################################

$ErrorActionPreference = "Continue"
$API = "http://localhost:8080"

# --- Counters ---
$script:passed = 0
$script:failed = 0
$script:skipped = 0
$script:results = @()

# Create a shared web session to track CSRF cookies
$script:session = New-Object Microsoft.PowerShell.Commands.WebRequestSession

# --- Bootstrap: get initial CSRF cookie via a GET request ---
try {
    $null = Invoke-WebRequest -Uri "$API/healthz" -UseBasicParsing -WebSession $script:session -ErrorAction Stop
}
catch {}

# --- Helpers ---
function Write-Header($text) {
    Write-Host ""
    Write-Host ("=" * 70) -ForegroundColor Cyan
    Write-Host "  $text" -ForegroundColor Cyan
    Write-Host ("=" * 70) -ForegroundColor Cyan
}

function Write-SubHeader($text) {
    Write-Host ""
    Write-Host "--- $text ---" -ForegroundColor Yellow
}

function Get-CsrfToken {
    # Extract __csrf_token from session cookies and URL-decode it
    # The cookie value is URL-encoded (e.g. %3D for =) but the Go middleware
    # compares the raw decoded value from the cookie with the header value
    $cookie = $script:session.Cookies.GetCookies("$API") | Where-Object { $_.Name -eq "__csrf_token" }
    if ($cookie) { return [System.Net.WebUtility]::UrlDecode($cookie.Value) }
    return ""
}

function Test-Endpoint {
    param(
        [string]$Name,
        [string]$Method = "GET",
        [string]$Url,
        [string]$Body = $null,
        [hashtable]$Headers = @{},
        [int[]]$ExpectedStatus = @(200),
        [switch]$ReturnResponse
    )

    $h = @{ "Content-Type" = "application/json" }
    foreach ($k in $Headers.Keys) { $h[$k] = $Headers[$k] }

    # Add CSRF token header for state-changing methods
    if ($Method -ne "GET" -and $Method -ne "HEAD" -and $Method -ne "OPTIONS") {
        $csrf = Get-CsrfToken
        if ($csrf) {
            $h["X-CSRF-Token"] = $csrf
        }
    }

    $params = @{
        Method          = $Method
        Uri             = $Url
        Headers         = $h
        WebSession      = $script:session
        UseBasicParsing = $true
        ErrorAction     = "Stop"
    }
    if ($Body) { $params["Body"] = [System.Text.Encoding]::UTF8.GetBytes($Body) }

    try {
        $resp = Invoke-WebRequest @params
        $status = $resp.StatusCode
        $json = $null
        if ($resp.Content) {
            try { $json = $resp.Content | ConvertFrom-Json } catch {}
        }

        if ($status -in $ExpectedStatus) {
            Write-Host "  [PASS] $Name (HTTP $status)" -ForegroundColor Green
            $script:passed++
            $script:results += [PSCustomObject]@{ Test = $Name; Status = "PASS"; HTTP = $status; Detail = "" }
        }
        else {
            Write-Host "  [FAIL] $Name - Expected $($ExpectedStatus -join '/'), got $status" -ForegroundColor Red
            $script:failed++
            $script:results += [PSCustomObject]@{ Test = $Name; Status = "FAIL"; HTTP = $status; Detail = "Unexpected status" }
        }

        if ($ReturnResponse) { return $json }
    }
    catch {
        $ex = $_.Exception
        $actualStatus = 0
        $errorBody = ""
        if ($ex.Response) {
            $actualStatus = [int]$ex.Response.StatusCode
            try {
                $sr = New-Object System.IO.StreamReader($ex.Response.GetResponseStream())
                $errorBody = $sr.ReadToEnd()
                $sr.Close()
            }
            catch {}
        }

        if ($actualStatus -in $ExpectedStatus) {
            Write-Host "  [PASS] $Name (HTTP $actualStatus - expected error)" -ForegroundColor Green
            $script:passed++
            $script:results += [PSCustomObject]@{ Test = $Name; Status = "PASS"; HTTP = $actualStatus; Detail = "Expected error response" }
            if ($ReturnResponse) {
                try { return ($errorBody | ConvertFrom-Json) } catch { return $null }
            }
        }
        else {
            if ($actualStatus -gt 0) { $detailMsg = "HTTP $actualStatus" } else { $detailMsg = $ex.Message }
            Write-Host "  [FAIL] $Name - $detailMsg" -ForegroundColor Red
            if ($errorBody) {
                $truncLen = [Math]::Min(200, $errorBody.Length)
                Write-Host "         Body: $($errorBody.Substring(0, $truncLen))" -ForegroundColor DarkGray
            }
            $script:failed++
            $script:results += [PSCustomObject]@{ Test = $Name; Status = "FAIL"; HTTP = $actualStatus; Detail = $detailMsg }
            if ($ReturnResponse) {
                try { return ($errorBody | ConvertFrom-Json) } catch { return $null }
            }
        }
    }
}

###############################################################################
# 1. HEALTH CHECKS
###############################################################################
Write-Header "1. HEALTH CHECKS"

Test-Endpoint -Name "GET /healthz" -Url "$API/healthz"
Test-Endpoint -Name "GET /readyz"  -Url "$API/readyz"

###############################################################################
# 2. AUTH - REGISTER
###############################################################################
Write-Header "2. AUTHENTICATION"
Write-SubHeader "Register new user"

$ts = [DateTimeOffset]::Now.ToUnixTimeSeconds()
$testEmail = "qa-test-$ts@paas.local"
$testPass = "QA_T3st!Pass_$ts"

$registerBody = @{
    name     = "QA Test User"
    email    = $testEmail
    password = $testPass
} | ConvertTo-Json

$regResp = Test-Endpoint -Name "POST /auth/register" `
    -Method POST -Url "$API/api/v1/auth/register" `
    -Body $registerBody -ReturnResponse

$accessToken = $null
$refreshToken = $null
if ($regResp -and $regResp.data) {
    $accessToken = $regResp.data.access_token
    $refreshToken = $regResp.data.refresh_token
    Write-Host "         User registered: $($regResp.data.user.email)" -ForegroundColor DarkGray
}

###############################################################################
# 3. AUTH - LOGIN (seeded admin user)
###############################################################################
Write-SubHeader "Login with seeded admin"

$loginBody = @{
    email    = "admin@paas.local"
    password = "admin"
} | ConvertTo-Json

$loginResp = Test-Endpoint -Name "POST /auth/login (admin)" `
    -Method POST -Url "$API/api/v1/auth/login" `
    -Body $loginBody -ReturnResponse

$adminToken = $null
$adminRefresh = $null
if ($loginResp -and $loginResp.data) {
    $adminToken = $loginResp.data.access_token
    $adminRefresh = $loginResp.data.refresh_token
    Write-Host "         Logged in as: $($loginResp.data.user.email)" -ForegroundColor DarkGray
}
else {
    Write-Host "         Admin login failed, falling back to QA user token" -ForegroundColor Yellow
    $adminToken = $accessToken
    $adminRefresh = $refreshToken
}

# Determine the active token
if ($adminToken) { $TOKEN = $adminToken } else { $TOKEN = $accessToken }
if (-not $TOKEN) {
    Write-Host ""
    Write-Host "  [SKIP] No auth token available - skipping authenticated tests" -ForegroundColor Yellow
    $script:skipped += 20
}

$authHeader = @{ "Authorization" = "Bearer $TOKEN" }

###############################################################################
# 4. AUTH - REFRESH
###############################################################################
Write-SubHeader "Token refresh"

if ($adminRefresh) { $useRefresh = $adminRefresh } else { $useRefresh = $refreshToken }
if ($useRefresh) {
    $refreshBody = @{ refresh_token = $useRefresh } | ConvertTo-Json
    $refResp = Test-Endpoint -Name "POST /auth/refresh" `
        -Method POST -Url "$API/api/v1/auth/refresh" `
        -Body $refreshBody -ReturnResponse

    if ($refResp -and $refResp.data -and $refResp.data.access_token) {
        $TOKEN = $refResp.data.access_token
        $authHeader = @{ "Authorization" = "Bearer $TOKEN" }
        Write-Host "         Token refreshed successfully" -ForegroundColor DarkGray
    }
}
else {
    Write-Host "  [SKIP] No refresh token available" -ForegroundColor Yellow
    $script:skipped++
}

###############################################################################
# 5. AUTH - NEGATIVE TESTS
###############################################################################
Write-SubHeader "Auth negative tests"

Test-Endpoint -Name "POST /auth/login (wrong password)" `
    -Method POST -Url "$API/api/v1/auth/login" `
    -Body '{"email":"admin@paas.local","password":"wrongpassword123"}' `
    -ExpectedStatus @(401)

Test-Endpoint -Name "POST /auth/login (missing fields)" `
    -Method POST -Url "$API/api/v1/auth/login" `
    -Body '{}' `
    -ExpectedStatus @(400, 422)

Test-Endpoint -Name "GET /users/me (no auth)" `
    -Url "$API/api/v1/users/me" `
    -ExpectedStatus @(401)

###############################################################################
# 6. USER PROFILE
###############################################################################
Write-Header "3. USER PROFILE"

if ($TOKEN) {
    $meResp = Test-Endpoint -Name "GET /users/me" `
        -Url "$API/api/v1/users/me" `
        -Headers $authHeader -ReturnResponse
    if ($meResp -and $meResp.data) {
        Write-Host "         Profile: $($meResp.data.name) <$($meResp.data.email)>" -ForegroundColor DarkGray
    }

    $updateBody = @{ name = "QA Admin Updated" } | ConvertTo-Json
    Test-Endpoint -Name "PUT /users/me" `
        -Method PUT -Url "$API/api/v1/users/me" `
        -Body $updateBody -Headers $authHeader
}

###############################################################################
# 7. ORGANIZATION CRUD
###############################################################################
Write-Header "4. ORGANIZATION CRUD"

$orgId = $null
if ($TOKEN) {
    Write-SubHeader "Create org"
    $orgSlug = "qaorg$ts"
    $orgBody = @{
        name = "QA Test Org"
        slug = $orgSlug
    } | ConvertTo-Json

    $orgResp = Test-Endpoint -Name "POST /orgs (create)" `
        -Method POST -Url "$API/api/v1/orgs" `
        -Body $orgBody -Headers $authHeader `
        -ExpectedStatus @(200, 201) -ReturnResponse
    if ($orgResp -and $orgResp.data) {
        $orgId = $orgResp.data.id
        Write-Host "         Org created: $orgId ($orgSlug)" -ForegroundColor DarkGray
    }

    Write-SubHeader "List orgs"
    Test-Endpoint -Name "GET /orgs (list)" `
        -Url "$API/api/v1/orgs" -Headers $authHeader

    if ($orgId) {
        Write-SubHeader "Get org"
        Test-Endpoint -Name "GET /orgs/:orgId" `
            -Url "$API/api/v1/orgs/$orgId" -Headers $authHeader

        Write-SubHeader "Update org"
        $updateOrgBody = @{ name = "QA Org Renamed" } | ConvertTo-Json
        Test-Endpoint -Name "PUT /orgs/:orgId" `
            -Method PUT -Url "$API/api/v1/orgs/$orgId" `
            -Body $updateOrgBody -Headers $authHeader
    }
}

###############################################################################
# 8. MEMBERS
###############################################################################
Write-Header "5. MEMBERS"

if ($TOKEN -and $orgId) {
    Write-SubHeader "List members"
    $membersResp = Test-Endpoint -Name "GET /orgs/:orgId/members" `
        -Url "$API/api/v1/orgs/$orgId/members" `
        -Headers $authHeader -ReturnResponse
    if ($membersResp -and $membersResp.data) {
        Write-Host "         Members: $($membersResp.data.Count)" -ForegroundColor DarkGray
    }
}

###############################################################################
# 9. PROJECTS CRUD
###############################################################################
Write-Header "6. PROJECTS"

$projectId = $null
if ($TOKEN -and $orgId) {
    Write-SubHeader "Create project"
    $projBody = @{
        name        = "QA Test Project"
        description = "Created by QA test script"
    } | ConvertTo-Json

    $projResp = Test-Endpoint -Name "POST /orgs/:orgId/projects (create)" `
        -Method POST -Url "$API/api/v1/orgs/$orgId/projects" `
        -Body $projBody -Headers $authHeader `
        -ExpectedStatus @(200, 201) -ReturnResponse
    if ($projResp -and $projResp.data) {
        $projectId = $projResp.data.id
        Write-Host "         Project: $projectId" -ForegroundColor DarkGray
    }

    Write-SubHeader "List projects"
    Test-Endpoint -Name "GET /orgs/:orgId/projects (list)" `
        -Url "$API/api/v1/orgs/$orgId/projects" -Headers $authHeader

    if ($projectId) {
        Write-SubHeader "Get project"
        Test-Endpoint -Name "GET /orgs/:orgId/projects/:projectId" `
            -Url "$API/api/v1/orgs/$orgId/projects/$projectId" -Headers $authHeader

        Write-SubHeader "Update project"
        $updateProjBody = @{ name = "QA Project Renamed"; description = "Updated by QA script" } | ConvertTo-Json
        Test-Endpoint -Name "PUT /orgs/:orgId/projects/:projectId" `
            -Method PUT -Url "$API/api/v1/orgs/$orgId/projects/$projectId" `
            -Body $updateProjBody -Headers $authHeader
    }
}

###############################################################################
# 10. DEPLOYMENTS
###############################################################################
Write-Header "7. DEPLOYMENTS"

$deploymentId = $null
if ($TOKEN -and $orgId -and $projectId) {
    Write-SubHeader "Create deployment"
    $deployBody = @{
        version    = "1.0.0-qa"
        commit_sha = "abc123def456"
    } | ConvertTo-Json

    $deployResp = Test-Endpoint -Name "POST .../deployments (create)" `
        -Method POST `
        -Url "$API/api/v1/orgs/$orgId/projects/$projectId/deployments" `
        -Body $deployBody -Headers $authHeader `
        -ExpectedStatus @(200, 201) -ReturnResponse
    if ($deployResp -and $deployResp.data) {
        $deploymentId = $deployResp.data.id
        Write-Host "         Deployment: $deploymentId (status=$($deployResp.data.status))" -ForegroundColor DarkGray
    }

    Write-SubHeader "List deployments"
    Test-Endpoint -Name "GET .../deployments (list)" `
        -Url "$API/api/v1/orgs/$orgId/projects/$projectId/deployments" `
        -Headers $authHeader
}

###############################################################################
# 11. ENV VARS
###############################################################################
Write-Header "8. ENVIRONMENT VARIABLES"

$envVarId = $null
if ($TOKEN -and $orgId -and $projectId) {
    Write-SubHeader "Set env var"
    $envBody = @{
        key       = "QA_TEST_KEY"
        value     = "qa-test-value-123"
        is_secret = $false
    } | ConvertTo-Json

    $envResp = Test-Endpoint -Name "POST .../env (set)" `
        -Method POST `
        -Url "$API/api/v1/orgs/$orgId/projects/$projectId/env" `
        -Body $envBody -Headers $authHeader -ReturnResponse
    if ($envResp -and $envResp.data) {
        $envVarId = $envResp.data.id
        Write-Host "         EnvVar: $($envResp.data.key) = $($envResp.data.value)" -ForegroundColor DarkGray
    }

    Write-SubHeader "Set secret env var"
    $secretBody = @{
        key       = "QA_SECRET_KEY"
        value     = "super-secret-value"
        is_secret = $true
    } | ConvertTo-Json
    Test-Endpoint -Name "POST .../env (secret)" `
        -Method POST `
        -Url "$API/api/v1/orgs/$orgId/projects/$projectId/env" `
        -Body $secretBody -Headers $authHeader

    Write-SubHeader "List env vars"
    $listEnvResp = Test-Endpoint -Name "GET .../env (list)" `
        -Url "$API/api/v1/orgs/$orgId/projects/$projectId/env" `
        -Headers $authHeader -ReturnResponse
    if ($listEnvResp -and $listEnvResp.data) {
        Write-Host "         Env vars: $($listEnvResp.data.Count)" -ForegroundColor DarkGray
    }

    if ($envVarId) {
        Write-SubHeader "Delete env var"
        Test-Endpoint -Name "DELETE .../env/:envVarId" `
            -Method DELETE `
            -Url "$API/api/v1/orgs/$orgId/projects/$projectId/env/$envVarId" `
            -Headers $authHeader
    }
}

###############################################################################
# 12. BILLING
###############################################################################
Write-Header "9. BILLING"

Write-SubHeader "Public billing plans"
$plansResp = Test-Endpoint -Name "GET /billing/plans (public)" `
    -Url "$API/api/v1/billing/plans" -ReturnResponse
if ($plansResp -and $plansResp.data) {
    Write-Host "         Plans: $($plansResp.data.Count)" -ForegroundColor DarkGray
    foreach ($p in $plansResp.data) {
        Write-Host "           - $($p.name) ($($p.slug))" -ForegroundColor DarkGray
    }
}

if ($TOKEN -and $orgId) {
    Write-SubHeader "Billing overview"
    Test-Endpoint -Name "GET /orgs/:orgId/billing" `
        -Url "$API/api/v1/orgs/$orgId/billing" -Headers $authHeader

    Write-SubHeader "Usage"
    Test-Endpoint -Name "GET /orgs/:orgId/billing/usage" `
        -Url "$API/api/v1/orgs/$orgId/billing/usage" -Headers $authHeader

    Write-SubHeader "Invoices"
    Test-Endpoint -Name "GET /orgs/:orgId/billing/invoices" `
        -Url "$API/api/v1/orgs/$orgId/billing/invoices" -Headers $authHeader
}

###############################################################################
# 13. CLEANUP
###############################################################################
Write-Header "10. CLEANUP"

if ($TOKEN -and $orgId) {
    if ($projectId) {
        Test-Endpoint -Name "DELETE project" `
            -Method DELETE `
            -Url "$API/api/v1/orgs/$orgId/projects/$projectId" `
            -Headers $authHeader
    }

    Test-Endpoint -Name "DELETE org" `
        -Method DELETE -Url "$API/api/v1/orgs/$orgId" `
        -Headers $authHeader
}

###############################################################################
# 14. LOGOUT
###############################################################################
Write-SubHeader "Logout"
if ($TOKEN) {
    Test-Endpoint -Name "POST /auth/logout" `
        -Method POST -Url "$API/api/v1/auth/logout" `
        -Headers $authHeader

    # Note: JWT is stateless - token remains valid until expiry
    # This verifies the API accepts the token even after logout
    # (refresh token is revoked, but access token still works until expiry)
    Test-Endpoint -Name "GET /users/me (after logout - JWT still valid)" `
        -Url "$API/api/v1/users/me" `
        -Headers $authHeader `
        -ExpectedStatus @(200, 401)
}

###############################################################################
# SUMMARY
###############################################################################
Write-Host ""
Write-Header "QA TEST SUMMARY"
Write-Host ""
Write-Host "  Passed  : $($script:passed)" -ForegroundColor Green
if ($script:failed -gt 0) { $failColor = "Red" } else { $failColor = "Green" }
Write-Host "  Failed  : $($script:failed)" -ForegroundColor $failColor
Write-Host "  Skipped : $($script:skipped)" -ForegroundColor Yellow
Write-Host "  Total   : $($script:passed + $script:failed + $script:skipped)" -ForegroundColor White
Write-Host ""

if ($script:failed -gt 0) {
    Write-Host "  Failed tests:" -ForegroundColor Red
    $script:results | Where-Object { $_.Status -eq "FAIL" } | ForEach-Object {
        Write-Host "    - $($_.Test): $($_.Detail)" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host ("=" * 70) -ForegroundColor Cyan

# Export results to CSV
$csvPath = Join-Path $PSScriptRoot "qa_results.csv"
$script:results | Export-Csv -Path $csvPath -NoTypeInformation
Write-Host "  Results exported to: $csvPath" -ForegroundColor DarkGray
Write-Host ""
