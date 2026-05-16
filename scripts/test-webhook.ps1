# Mock webhook: logs POST body and returns 200
$listener = [System.Net.HttpListener]::new()
$listener.Prefixes.Add('http://127.0.0.1:8080/')
$listener.Start()
Write-Host "Webhook mock listening on http://127.0.0.1:8080/"

while ($listener.IsListening) {
    $context = $listener.GetContext()
    $request = $context.Request
    $response = $context.Response

    $reader = [System.IO.StreamReader]::new($request.InputStream, $request.ContentEncoding)
    $body = $reader.ReadToEnd()
    $reader.Close()

    Write-Host ""
    Write-Host "=== $($request.HttpMethod) $($request.Url.PathAndQuery) ==="
    Write-Host $body
    Write-Host "=== end ==="

    $buffer = [System.Text.Encoding]::UTF8.GetBytes('ok')
    $response.StatusCode = 200
    $response.ContentLength64 = $buffer.Length
    $response.OutputStream.Write($buffer, 0, $buffer.Length)
    $response.OutputStream.Close()
}
