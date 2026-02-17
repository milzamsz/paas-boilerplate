
$api = "http://localhost:8080"
$origin = "http://localhost:3000"

# 1. Get CSRF Token and Cookies
$response = Invoke-WebRequest -Uri "$api/healthz" -SessionVariable session
$csrfCookie = $session.Cookies.GetCookies($api) | Where-Object { $_.Name -eq "__csrf_token" }
$csrfToken = [System.Net.WebUtility]::UrlDecode($csrfCookie.Value)

Write-Host "CSRF Token: $csrfToken"

# 2. Attempt Login with Origin Header
$loginUrl = "$api/api/v1/auth/login"
$body = @{
    email    = "admin@paas.local"
    password = "admin"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-WebRequest -Uri $loginUrl -Method Post -Body $body -ContentType "application/json" -WebSession $session -Headers @{
        "X-CSRF-Token" = $csrfToken
        "Origin"       = $origin
    }
    
    Write-Host "Login Status: $($loginResponse.StatusCode)"
    Write-Host "Access-Control-Allow-Credentials: $($loginResponse.Headers['Access-Control-Allow-Credentials'])"
    Write-Host "Access-Control-Allow-Origin: $($loginResponse.Headers['Access-Control-Allow-Origin'])"

}
catch {
    Write-Host "Login Request Failed: $_"
    # Even if it fails (e.g. 401), we want to see CORS headers
    if ($_.Exception.Response) {
        Write-Host "Status: $($_.Exception.Response.StatusCode)"
        Write-Host "Access-Control-Allow-Credentials: $($_.Exception.Response.Headers['Access-Control-Allow-Credentials'])"
        Write-Host "Access-Control-Allow-Origin: $($_.Exception.Response.Headers['Access-Control-Allow-Origin'])"
    }
}
