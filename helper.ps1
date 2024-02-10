go mod init github.com/kcloutie/er
go mod tidy


# =====================================================================================================
# Example
# =====================================================================================================

# Uncomment the example you want to run in the attributes section
# UnComment "test" = "all"  to run all the tests 

$dataJson = Get-Content -Path "test/testdata/data-example.json" -Raw | Out-String
$enc = [system.Text.Encoding]::UTF8
$data = $enc.GetBytes($dataJson) 

$Payload = @{
  ID = "1234"
  Attributes = @{
    "test" = "all"
    # "test" = "github"
    # "test" = "email"
    # "test" = "powershell"
    # "test" = "webex"
    # "test" = "webhook"
  }
  Data = $data
}
$Headers = @{
  "X-Cloud-Trace-Context" = (New-Guid | Select-Object -ExpandProperty Guid)
}
Invoke-WebRequest -Uri http://localhost:8080/api/v1/pubsub -Method Post -Body ($Payload | ConvertTo-Json -Compress) -ContentType "application/json" -Headers $Headers



export KUBECONFIG=/home/kcloutie/.kube/config.kind
kubectl logs deployment/er-controller-manager -n er
kubectl get deployment er-controller-manager -n er

kubectl delete deployment er-controller-manager -n er
make rdev
kubectl logs deployment/ingress-nginx-controller  -n ingress-nginx

kubectl get events --sort-by='.metadata.creationTimestamp' -A
kubectl get pods -n er