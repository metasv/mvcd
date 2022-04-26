// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/metasv/bsvlog"
	"github.com/metasv/mvcd/blockchain"
	"github.com/metasv/mvcd/blockchain/indexers"
	"github.com/metasv/mvcd/database"
	"github.com/metasv/mvcd/limits"
)

const (
	// blockDbNamePrefix is the prefix for the bsvd block database.
	blockDbNamePrefix = "blocks"
)

var (
	cfg *config
	log bsvlog.Logger
)

// loadBlockDB opens the block database and returns a handle to it.
func loadBlockDB() (database.DB, error) {
	// The database name is based on the database type.
	dbName := blockDbNamePrefix + "_" + cfg.DbType
	dbPath := filepath.Join(cfg.DataDir, dbName)

	log.Infof("Loading block database from '%s'", dbPath)
	db, err := database.Open(cfg.DbType, dbPath, activeNetParams.Net)
	if err != nil {
		// Return the error if it's not because the database doesn't
		// exist.
		if dbErr, ok := err.(database.Error); !ok || dbErr.ErrorCode !=
			database.ErrDbDoesNotExist {

			return nil, err
		}

		// Create the db if it does not exist.
		err = os.MkdirAll(cfg.DataDir, 0700)
		if err != nil {
			return nil, err
		}
		db, err = database.Create(cfg.DbType, dbPath, activeNetParams.Net)
		if err != nil {
			return nil, err
		}
	}

	log.Info("Block database loaded")
	return db, nil
}

// realMain is the real main function for the utility.  It is necessary to work
// around the fact that deferred functions do not run when os.Exit() is called.
func realMain() error {
	// Load configuration and parse command line.
	tcfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	cfg = tcfg

	// Setup logging.
	backendLogger := bsvlog.NewBackend(os.Stdout)
	defer os.Stdout.Sync()
	log = backendLogger.Logger("MAIN")
	database.UseLogger(backendLogger.Logger("BCDB"))
	blockchain.UseLogger(backendLogger.Logger("CHAN"))
	indexers.UseLogger(backendLogger.Logger("INDX"))

	// Load the block database.
	db, err := loadBlockDB()
	if err != nil {
		log.Errorf("Failed to load database: %v", err)
		return err
	}
	defer db.Close()

	var utxoWriter io.Writer
	if cfg.OutFile != "" {
		utxoFile, err := os.Create(cfg.OutFile)
		if err != nil {
			log.Errorf("Unable to create file at: %s", cfg.OutFile)
			return err
		}
		defer utxoFile.Close()

		utxoWriter = utxoFile
	}

	// Perform the import asynchronously.  This allows blocks to be
	// processed and read in parallel.  The results channel returned from
	// Import contains the statistics about the import including an error
	// if something went wrong.
	log.Info("Starting Utxo hash calculation")
	utxoHash, totalSize, err := CalcUtxoSet(db, cfg.BlockHeight, utxoWriter)
	if err != nil {
		log.Errorf("%v", err)
		return err
	}

	log.Infof("Utxo set ECMH hash at height %d: %s (serialized size %d)", cfg.BlockHeight, utxoHash.String(), totalSize)
	return nil
}

func main() {
	// Use all processor cores and up some limits.
	runtime.GOMAXPROCS(runtime.NumCPU())
	if err := limits.SetLimits(); err != nil {
		os.Exit(1)
	}

	// Work around defer not working after os.Exit()
	if err := realMain(); err != nil {
		os.Exit(1)
	}
}
