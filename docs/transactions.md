# Transactions

Merhongo provides support for MongoDB transactions, allowing you to perform multiple database operations atomically. This ensures that either all operations succeed or none of them are applied, maintaining data consistency.

## Understanding Transactions

Transactions are essential when you need to execute multiple related operations as a single atomic unit. For example, transferring money between two accounts should either complete entirely or fail entirely to avoid inconsistent states.

## Prerequisites

To use transactions in MongoDB:

1. You must be using a replica set or a sharded cluster
2. MongoDB server version must be 4.0 or higher

## Basic Transaction Usage

Merhongo provides the `ExecuteTransaction` method on the client to execute operations in a transaction:

```go
import (
    "context"
    "github.com/isimtekin/merhongo"
    "github.com/isimtekin/merhongo/connection"
    "go.mongodb.org/mongo-driver/mongo"
)

func main() {
    // Connect to MongoDB
    client, err := merhongo.Connect("mongodb://localhost:27017", "mydatabase")
    if err != nil {
        log.Fatalf("Failed to connect to MongoDB: %v", err)
    }
    defer merhongo.Disconnect()
    
    // Get connection client
    conn := merhongo.GetConnection()
    
    // Execute operations in a transaction
    err = conn.ExecuteTransaction(context.Background(), func(sc mongo.SessionContext) error {
        // Perform operations within the transaction
        
        // Example: Create a user
        err := userModel.Create(sc, user)
        if err != nil {
            return err // This will cause the transaction to abort
        }
        
        // Example: Create related profile
        err = profileModel.Create(sc, profile)
        if err != nil {
            return err // This will cause the transaction to abort
        }
        
        // If we reach here, both operations succeeded
        return nil // This will commit the transaction
    })
    
    if err != nil {
        log.Fatalf("Transaction failed: %v", err)
    } else {
        log.Println("Transaction completed successfully")
    }
}
```

## Transaction Behavior

When you use `ExecuteTransaction`:

1. A transaction is started automatically
2. Your function is executed within the transaction context
3. If your function returns an error, the transaction is aborted
4. If your function returns nil, the transaction is committed
5. If an error occurs during commit, it's returned from `ExecuteTransaction`

## Using Models in Transactions

When using models within a transaction, you need to pass the session context:

```go
err = conn.ExecuteTransaction(ctx, func(sc mongo.SessionContext) error {
    // Create first document
    user := &User{
        Username: "john_doe",
        Email:    "john@example.com",
    }
    err := userModel.Create(sc, user) // Use session context (sc)
    if err != nil {
        return err
    }
    
    // Create second document
    profile := &Profile{
        UserID:      user.ID,
        DisplayName: "John Doe",
        Bio:         "Software developer",
    }
    err = profileModel.Create(sc, profile) // Use session context (sc)
    if err != nil {
        return err
    }
    
    return nil
})
```

All Merhongo model methods accept a `context.Context`, which can be a `mongo.SessionContext` for transactions.

## Example: Money Transfer

Here's a real-world example of using transactions for a money transfer:

```go
type Account struct {
    ID      primitive.ObjectID `bson:"_id,omitempty"`
    UserID  primitive.ObjectID `bson:"userId"`
    Balance float64            `bson:"balance"`
}

func TransferMoney(from, to primitive.ObjectID, amount float64) error {
    conn := merhongo.GetConnection()
    ctx := context.Background()
    
    return conn.ExecuteTransaction(ctx, func(sc mongo.SessionContext) error {
        // 1. Find the source account
        var fromAccount Account
        err := accountModel.FindOne(sc, bson.M{"_id": from}, &fromAccount)
        if err != nil {
            return errors.Wrap(err, "failed to find source account")
        }
        
        // 2. Check sufficient balance
        if fromAccount.Balance < amount {
            return errors.WithDetails(errors.ErrValidation, "insufficient funds")
        }
        
        // 3. Find the destination account
        var toAccount Account
        err = accountModel.FindOne(sc, bson.M{"_id": to}, &toAccount)
        if err != nil {
            return errors.Wrap(err, "failed to find destination account")
        }
        
        // 4. Update the source account
        err = accountModel.UpdateById(sc, from.Hex(), bson.M{
            "$inc": bson.M{"balance": -amount},
        })
        if err != nil {
            return errors.Wrap(err, "failed to update source account")
        }
        
        // 5. Update the destination account
        err = accountModel.UpdateById(sc, to.Hex(), bson.M{
            "$inc": bson.M{"balance": amount},
        })
        if err != nil {
            return errors.Wrap(err, "failed to update destination account")
        }
        
        // 6. Create a transaction record
        transfer := &Transfer{
            FromAccount: from,
            ToAccount:   to,
            Amount:      amount,
            Timestamp:   time.Now(),
        }
        err = transferModel.Create(sc, transfer)
        if err != nil {
            return errors.Wrap(err, "failed to create transfer record")
        }
        
        return nil
    })
}
```

## Transaction Limitations

1. **Timeout**: Transactions have a default 60-second timeout. Long-running transactions might timeout.
2. **Lock Contention**: Transactions can increase lock contention and reduce concurrency.
3. **Performance Impact**: Transactions have overhead and might impact performance. Use them only when necessary.

## Best Practices

1. **Keep transactions short and simple**: Minimize the operations within a transaction.
2. **Handle errors properly**: Always check error returns and abort transactions on failure.
3. **Specify read/write concerns**: For advanced use cases, provide appropriate read and write concerns.
4. **Consider retries for transient errors**: Some transaction errors are temporary and can be resolved by retrying.
5. **Use transactions only when necessary**: For single operations, transactions add unnecessary overhead.
6. **Test with a replica set**: Ensure your application is tested with a proper MongoDB deployment that supports transactions.