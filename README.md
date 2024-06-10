# go-txmanager
Transaction Manager with gorm. Supports gorm v1 and v2.

## How To Use

### Prerequisite on repository

make sure to pass context on a repository, and check txConn like
```go
var tx *gorm.DB
if txConn := txmanager.GetTxConn(ctx); txConn != nil {
    tx = txConn
}
if tx == nil {
    tx = p.db.Begin()
    defer func() {
        if err != nil {
            tx.Rollback()
            return
        }
        tx.Commit()
    }()
}
```
this will check current tx on passing context

### Making transaction with TxManager
first, you need to start the TxManager
```
txManager := txmanager.StartTxManager(db)
```
and then make a transaction function, which form like this (just an example),
you need to pass context to repository
```go
transaction := func(ctx context.Context) error {
    err := repoA.Update(ctx, id, model)
    if err != nil {
        return err
    }

    err := repoB.Update(ctx, id, model)
    if err != nil {
        return err
    }
    return nil
}
```
and to execute the transaction
```
err := txManager.WithTransaction(context, transaction)
```
also you can find the example on `txmanager_integration_test.go`

## Testing

### Run Test
Run the following command to start the local test mysql 
```
$ make test-infra-up
```

Run the test
```
$ make test
```

Stop the test mysql when finish testing
```
$ make test-infra-down
```

