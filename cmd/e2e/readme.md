## Run all tests
```
go test -timeout 30s -v
```

## Run Only Individual E2E Tests
```
go test -timeout 30s -run '^Test\w*Command' -v
```

## Run Only Flow E2E Tests
```
go test -timeout 30s -run '^Test\w*Flow' -v  
```