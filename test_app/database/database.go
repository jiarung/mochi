package database

import (
    "errors"
    "fmt"
//    "net/url"
//    "os"
    "reflect"
    "regexp"
   "runtime/debug"
    "strings"
    "sync"
    "time"

    "github.com/jinzhu/gorm"
 //   _ "gorm.io/driver/postgres"
    _ "github.com/jinzhu/gorm/dialects/postgres"


	"github.com/jiarung/mochi/test_app/api/models"
)

/*
// DBApp is an interface for different database applications.
type DBApp interface {
	// Models returns the models for a given database app.
	Models() []interface{}

	// IsEmpty check if a given database is empty.
	IsEmpty(db *gorm.DB) bool

    // Reset reset db.
    Reset(db *gorm.DB) error
}
*/

// DBApp is the database application for cobinhood.
type DBApp struct {
}

// Models returns the models for a given database app.
func (e *DBApp) Models() []interface{} {
	return models.AllModels
}

// IsEmpty check if a given database is empty.
func (e *DBApp) IsEmpty(db *gorm.DB) bool {
	return !db.HasTable("user")
}

// Reset is executed db reset.
func (e *DBApp) Reset(tx *gorm.DB) error {
	return nil
}

// Host specifies the database host.
type Host string

// Host enums.
const (
    Default Host = "default"
)


// DB defines db connections.
type DB struct {
    conns    map[Host]*gorm.DB
    connLock sync.Mutex
}

var defaultDB DB

// conn configures given params in shared flow and returns an ORM DB
// instance with same functionalities and features.
// In this package, we create global DB instance with shared params.
// For external package, e.g: recod, it connects to restored instance with
// different params.
func (d *DB) conn(dialect string, host Host) (db *gorm.DB, err error) {
    db, err = gorm.Open("postgres", dialect)
    //    db, err = gorm.Open(dialect, args)

    if err != nil {
        return
    }

    // Set database parameters.
    db.DB().SetMaxIdleConns(2)
    db.DB().SetMaxOpenConns(10)
    db.DB().SetConnMaxLifetime(3 * time.Minute)

    return
}

// Initialize initializes models.
// It only creates the connection instance, doesn't reset or migrate anything.
func (d *DB) Initialize(host Host) {
    d.connLock.Lock()
    defer d.connLock.Unlock()

    d.conns = make(map[Host]*gorm.DB)

    dialect := "host=localhost user=gorm password=gorm dbname=gorm port=15432 sslmode=disable"
    db, err := d.conn(dialect, Default)
    if err != nil {
        fmt.Printf("Dial failed %v", err)
		panic(err)
    }

    d.conns[Default] = db

    dbName := "test dcard"
    err = d.conns[Default].Exec("CREATE DATABASE " + dbName).Error
    if err != nil {
        fmt.Printf("create database: %v", err)
    }
    fmt.Printf("sdfdsf")

    for key, db := range d.conns {
        if db != nil {
            db.Close()
            delete(d.conns, key)
        }
    }

}

// Finalize closes db.
func Finalize() {
    defaultDB.Finalize()
}

// Finalize closes the database.
func (d *DB) Finalize() {
    d.connLock.Lock()
    defer d.connLock.Unlock()

    for _, db := range d.conns {
        if db != nil {
            db.Close()
        }
    }
}

// GetDB gets db from singleton.
func GetDB(target Host) *gorm.DB {
    return defaultDB.GetDB(target)
}

// GetDB returns the database handle.
func (d *DB) GetDB(target Host) *gorm.DB {
	d.connLock.Lock()
	ret, ok := d.conns[target]
	d.connLock.Unlock()

	if !ok {
		fmt.Printf("uninitialized target: %v", target)
	}

	return ret
}


// Reset resets database.
func (d *DB) Reset(db *gorm.DB, dbApp DBApp) {

    fmt.Printf("Resetting database ...")

    // Must make sure only reset() can calls dropAllDatabase().
    if !hasTables(db, dbApp) {
        dropAllDatabase(db, dbApp)
    } else {
        // Ignore schema creation and only insert default record.
        d.DeleteAllData(dbApp)
        fmt.Printf("already has table, ignored")

        err := Transaction(db, func(tx *gorm.DB) error {
            fmt.Printf("Running post reset hook ...")
            if tErr := InitializeTables(nil, tx); tErr != nil {
                return tErr
            }

            return nil
        })

        if err != nil {
            fmt.Printf("post reset: %v", err)
        }

        return
    }

    fmt.Printf("Creating models ...")

    err := Transaction(db, func(tx *gorm.DB) error {
        // Create all tables.
        for _, model := range dbApp.Models() {
            if err := tx.AutoMigrate(model).Error; err != nil {
                return err
            }
        }
        return nil
    })
    if err != nil {
        panic(err)
    }

    err = Transaction(db, func(tx *gorm.DB) error {
        fmt.Printf("Creating indices and constraints ...")
        for _, v := range dbApp.Models() {
            name := tx.NewScope(v).TableName()
            if err := CreateCustomIndices(tx, v, name); err != nil {
                return err
            }

            if err := CreateForeignKeyConstraintsSelf(tx, v, name); err != nil {
                return err
            }
        }

        fmt.Printf("Running post reset hook ...")

        if tErr := InitializeTables(nil, tx); tErr != nil {
            return tErr
        }

        return nil
    })
    if err != nil {
        panic(err)
    }

    fmt.Printf("Done")
}

// DeleteAllData with `DELETE` since TRUNCATE runs so slow in cockroachdb
// This method provides same functionality as `DropAllData`
func (d *DB) DeleteAllData(dbApp DBApp) {

    fmt.Printf("`DELETE` data in all tables")
    var db *gorm.DB
    d.connLock.Lock()
    defer d.connLock.Unlock()
    db = d.conns[Default]
	defer db.Close()

    tx := db.Begin()
    allModels := dbApp.Models()
    for i := 0; i < len(allModels); i++ {
        v := allModels[len(allModels)-1-i]
        scope := db.NewScope(v)
        tablename := scope.TableName()
        tx.Exec(fmt.Sprintf("DELETE FROM \"%v\"", tablename))
        for _, field := range scope.GetModelStruct().StructFields {
            if relation := field.Relationship; relation != nil {
                if handler := relation.JoinTableHandler; handler != nil {
                    tablename = handler.Table(tx)
                    tx.Exec(fmt.Sprintf("DELETE FROM \"%v\"", tablename))
                }
            }
        }
    }
    // We may call DeleteAllData() at SetupSuite and call Reset() at SetupTest.
    // It will cause TRUNCATE failed, so we ignore Commit().Error.
    tx.Commit()

    fmt.Printf("Done")
}

func hasTables(db *gorm.DB, app DBApp) bool {
    models := app.Models()

    type Result struct {
        TableName string
    }

    var results []Result
    err := db.Raw(`SELECT table_name
    FROM information_schema.tables
    WHERE table_type = 'BASE TABLE'
    AND table_schema = CURRENT_SCHEMA()`).Scan(&results).Error
    if err != nil {
        return false
    }

    if len(results) != len(models) {
        return false
    }

    for _, v := range models {
        tablename := db.NewScope(v).TableName()
        found := false
        for _, res := range results {
            if res.TableName == tablename {
                found = true
                break
            }
        }
        if !found {
            return false
        }
    }

    return true
}

func dropAllDatabase(db *gorm.DB, dbApp DBApp) {
    fmt.Printf("Dropping old database ...")
    tx := db.Begin()
    for _, model := range dbApp.Models() {
        sql := fmt.Sprintf("DROP TABLE IF EXISTS \"%s\" CASCADE",
            db.NewScope(model).TableName())
        if err := tx.Exec(sql).Error; err != nil {
            fmt.Printf("Exec '%s' failed. err: %s", sql, err)
        }
    }
    if err := tx.Commit().Error; err != nil {
        panic(err)
    }
}

// Transaction wraps the database transaction and to proper error handling.
func Transaction(db *gorm.DB, body func(*gorm.DB) error) (err error) {
	tx := db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("Transaction: Cannot open transaction %v", tx.Error)
	}

	// Handle runtime.Goexit. err won't be set when Goexit is called in body.
	var errDefault error
	errDefault = errors.New("init")
	err = errDefault

	// Error checking and panic safenet.
	defer func() {
		if err != nil {
			if err == errDefault {
				fmt.Printf("Transaction: rollback due to unexpected error")
			} else {
				fmt.Printf("Transaction: rollback due to error: %v", err)
			}
			if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
				panic(rollbackErr)
			}
		}

		if recovered := recover(); recovered != nil {
			fmt.Printf("Transaction: rollback due to panic: %v\n%s",
				recovered, string(debug.Stack()))

			err = tx.Rollback().Error
			if err != nil {
				fmt.Printf("Transaction: rollback failed: %v", err)
			}
			panic(recovered)
		}
	}()

	// Execute main body.
	if err = body(tx); err != nil {
		return err
	}

	err = tx.Commit().Error
	if err != nil {
		return err
	}

	return nil
}

// InitializeTables provides an interface for testing.
func InitializeTables(
    recordList []interface{}, tx *gorm.DB) error {
    if recordList == nil {
        return nil
    }
    for idx, record := range recordList {
        typeOfRecord := reflect.TypeOf(record)
        if typeOfRecord.Kind() != reflect.Ptr {
            return fmt.Errorf("typeOfRecord.Kind() != reflect.Ptr")
        }

        if result := tx.Create(record); result.Error != nil {
            return fmt.Errorf("tx.Create() (%d/%d) for %s failed! record(%+v), "+
                "result.Error(%s)",
                idx+1,
                len(recordList),
                typeOfRecord.Elem().Name(),
                record,
                result.Error)
        }
    }
    return nil
}

// CreateCustomIndices creates custom indices if model implements
// models.CustomIndexer.
func CreateCustomIndices(tx *gorm.DB, model interface{}, name string) error {
    if m, ok := model.(models.CustomIndexer); ok {
        for _, idx := range m.Indexes() {
            unique := ""
            extension := ""
            if idx.Unique {
                unique = "UNIQUE"
            }
            if 0 != len(idx.Type) {
                extension = "USING " + idx.Type
            }
            columns := strings.Join(idx.Fields, ",")
            idxStat := fmt.Sprintf(
                `CREATE %s INDEX IF NOT EXISTS %s_%s ON "%s" %s(%s) %s`,
                unique, name, idx.Name, name, extension, columns, idx.Condition)
            err := tx.Model(model).Exec(idxStat).Error
            if err != nil {
                return err
            }
        }
    }
    return nil
}

// CreateForeignKeyConstraints creates foreign key constraint if model
// implements models.ForeignKeyConstrainer.
func CreateForeignKeyConstraints(tx *gorm.DB, model interface{}) error {
    if m, ok := model.(models.ForeignKeyConstrainer); ok {
        for _, c := range m.ForeignKeyConstraints() {
            err := tx.Model(model).
                AddForeignKey(c.Field, c.Dest, c.OnDelete, c.OnDelete).Error
            if err != nil {
                return err
            }
        }
    }
    return nil
}

// CreateForeignKeyConstraintsSelf creates foreign key constraint if model
// implements models.ForeignKeyConstrainer.
func CreateForeignKeyConstraintsSelf(
    tx *gorm.DB, model interface{}, name string) error {
    if m, ok := model.(models.ForeignKeyConstrainer); ok {
        for _, c := range m.ForeignKeyConstraints() {
            keyName := buildForeignKeyName(name, c.Field, c.Dest)

            // result: ALTER TABLE table_name
            //         IF EXISTS ADD CONSTRAINT key_name FOREIGN KEY (field)
            //         REFERENCES dest ON DELETE onDelete ON UPDATE onUpdate;
            err := tx.Exec(
                fmt.Sprintf(`ALTER TABLE IF EXISTS "%s" ADD CONSTRAINT %s
                     FOREIGN KEY (%s) REFERENCES %s ON DELETE %s ON UPDATE %s`,
                    name, keyName,
                    c.Field, c.Dest, c.OnDelete, c.OnUpdate)).Error
            if err != nil {
                return err
            }
        }
    }
    return nil
}

// buildForeignKeyName is copy from gorm.
func buildForeignKeyName(tableName, field, dest string) string {
    keyName := fmt.Sprintf("%s_%s_%s_foreign", tableName, field, dest)
    keyName = regexp.MustCompile("[^a-zA-Z0-9]+").
        ReplaceAllString(keyName, "_")
    return keyName
}

/*
// Reset resets the entire database. It will:
// 1. Drop all database.
// 2. Do migration (contains initial schema & default recods).
func Reset(db *gorm.DB, dbApp DBApp, force bool) {
    defaultDB.Reset(db, dbApp, force)
}


func (d *DB) dialDB(host Host) *gorm.DB {
    var (
        args string
        db   *gorm.DB
        err  error
    )
    switch host {
    case Default:
        args = d.DbArgs
    case Master:
        args = d.DbMasterArgs
    case Statsd:
        args = d.DbStatsdArgs
    case Recod:
        args = d.DbRecodArgs
    }

    db, err = d.conn(d.Dialect, args, host)
    if err != nil {
        d.logger.Critical(err.Error())
    }
    return db
}

// Transaction wraps the database transaction and to proper error handling.
func Transaction(db *gorm.DB, body func(*gorm.DB) error) (err error) {
    tx := db.Begin()
    if tx.Error != nil {
        return fmt.Errorf("Transaction: Cannot open transaction %v", tx.Error)
    }

    // Handle runtime.Goexit. err won't be set when Goexit is called in body.
    var errDefault error
    errDefault = errors.New("init")
    err = errDefault

    // Error checking and panic safenet.
    defer func() {
        if err != nil {
            if err == errDefault {
                defaultDB.logger.
                    Warn("Transaction: rollback due to unexpected error")
            } else {
                defaultDB.logger.
                    Warn("Transaction: rollback due to error: %v", err)
            }
            if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
                panic(rollbackErr)
            }
        }

        if recovered := recover(); recovered != nil {
            defaultDB.fmt.Printf("Transaction: rollback due to panic: %v\n%s",
                recovered, string(debug.Stack()))

            err = tx.Rollback().Error
            if err != nil {
                defaultDB.fmt.Printf("Transaction: rollback failed: %v", err)
            }
            panic(recovered)
        }
    }()

    // Execute main body.
    if err = body(tx); err != nil {
        return err
    }

    err = tx.Commit().Error
    if err != nil {
        return err
    }

    return nil
}

// ShouldBeTransaction check the input `tx` is in a transaction or not.
// This is ugly, but it make sure the caller do the right thing.
func ShouldBeTransaction(tx *gorm.DB) bool {
    // clone to disable debug mode.
    cloneTx := tx.New()
    cloneTx.LogMode(false)

    // check the tx must be a transaction
    // it could be simply check by tx.db.(sqlDb) but gorm not expose this.
    dummyTx := cloneTx.Begin()
    if dummyTx.Error != gorm.ErrCantStartTransaction {
        if dummyTx.Error == nil {
            dummyTx.Rollback()
        }
        return false
    }

    return true
}

*/
