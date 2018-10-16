package db

import (
	"database/sql"
	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/op/go-logging"
	"github.com/textileio/textile-go/repo"
	"path"
	"sync"
)

var log = logging.MustGetLogger("db")

type SQLiteDatastore struct {
	config             repo.ConfigStore
	profile            repo.ProfileStore
	threads            repo.ThreadStore
	devices            repo.DeviceStore
	peers              repo.PeerStore
	blocks             repo.BlockStore
	notifications      repo.NotificationStore
	cafeSessions       repo.CafeSessionStore
	cafeRequests       repo.CafeRequestStore
	cafeNonces         repo.CafeNonceStore
	cafeAccounts       repo.CafeAccountStore
	cafeAccountThreads repo.CafeAccountThreadStore
	db                 *sql.DB
	lock               *sync.Mutex
}

func Create(repoPath, pin string) (*SQLiteDatastore, error) {
	var dbPath string
	dbPath = path.Join(repoPath, "datastore", "mainnet.db")
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	if pin != "" {
		p := "pragma key='" + pin + "';"
		conn.Exec(p)
	}
	mux := new(sync.Mutex)
	sqliteDB := &SQLiteDatastore{
		config:             NewConfigStore(conn, mux, dbPath),
		profile:            NewProfileStore(conn, mux),
		threads:            NewThreadStore(conn, mux),
		devices:            NewDeviceStore(conn, mux),
		peers:              NewPeerStore(conn, mux),
		blocks:             NewBlockStore(conn, mux),
		notifications:      NewNotificationStore(conn, mux),
		cafeSessions:       NewCafeSessionStore(conn, mux),
		cafeRequests:       NewCafeRequestStore(conn, mux),
		cafeNonces:         NewCafeNonceStore(conn, mux),
		cafeAccounts:       NewCafeAccountStore(conn, mux),
		cafeAccountThreads: NewCafeAccountThreadStore(conn, mux),
		db:                 conn,
		lock:               mux,
	}

	return sqliteDB, nil
}

func (d *SQLiteDatastore) Ping() error {
	return d.db.Ping()
}

func (d *SQLiteDatastore) Close() {
	d.db.Close()
}

func (d *SQLiteDatastore) Config() repo.ConfigStore {
	return d.config
}

func (d *SQLiteDatastore) Profile() repo.ProfileStore {
	return d.profile
}

func (d *SQLiteDatastore) Threads() repo.ThreadStore {
	return d.threads
}

func (d *SQLiteDatastore) Devices() repo.DeviceStore {
	return d.devices
}

func (d *SQLiteDatastore) Peers() repo.PeerStore {
	return d.peers
}

func (d *SQLiteDatastore) Blocks() repo.BlockStore {
	return d.blocks
}

func (d *SQLiteDatastore) Notifications() repo.NotificationStore {
	return d.notifications
}

func (d *SQLiteDatastore) CafeSessions() repo.CafeSessionStore {
	return d.cafeSessions
}

func (d *SQLiteDatastore) CafeRequests() repo.CafeRequestStore {
	return d.cafeRequests
}

func (d *SQLiteDatastore) CafeNonces() repo.CafeNonceStore {
	return d.cafeNonces
}

func (d *SQLiteDatastore) CafeAccounts() repo.CafeAccountStore {
	return d.cafeAccounts
}

func (d *SQLiteDatastore) CafeAccountThreads() repo.CafeAccountThreadStore {
	return d.cafeAccountThreads
}

func (d *SQLiteDatastore) Copy(dbPath string, password string) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	var cp string
	stmt := "select name from sqlite_master where type='table'"
	rows, err := d.db.Query(stmt)
	if err != nil {
		log.Errorf("error in copy: %s", err)
		return err
	}
	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		tables = append(tables, name)
	}
	if password == "" {
		cp = `attach database '` + dbPath + `' as plaintext key '';`
		for _, name := range tables {
			cp = cp + "insert into plaintext." + name + " select * from main." + name + ";"
		}
	} else {
		cp = `attach database '` + dbPath + `' as encrypted key '` + password + `';`
		for _, name := range tables {
			cp = cp + "insert into encrypted." + name + " select * from main." + name + ";"
		}
	}
	_, err = d.db.Exec(cp)
	if err != nil {
		return err
	}
	return nil
}

func (d *SQLiteDatastore) InitTables(password string) error {
	return initDatabaseTables(d.db, password)
}

func initDatabaseTables(db *sql.DB, pin string) error {
	var sqlStmt string
	if pin != "" {
		sqlStmt = "PRAGMA key = '" + pin + "';"
	}
	sqlStmt += `
	create table config (key text primary key not null, value blob);
    create table profile (key text primary key not null, value blob);
    create table threads (id text primary key not null, name text not null, sk blob not null, head text not null);
    create table devices (id text primary key not null, name text not null);
    create table peers (row text primary key not null, id text not null, pk blob not null, threadId text not null);
    create unique index peer_threadId_id on peers (threadId, id);
    create table blocks (id text primary key not null, date integer not null, parents text not null, threadId text not null, authorPk text not null, type integer not null, dataId text, dataKeyCipher blob, dataCaptionCipher blob, dataUsernameCipher blob, dataMetadataCipher blob);
    create index block_dataId on blocks (dataId);
    create index block_threadId_type_date on blocks (threadId, type, date);
    create table notifications (id text primary key not null, date integer not null, actorId text not null, actorUsername text not null, subject text not null, subjectId text not null, blockId text, dataId text, type integer not null, body text not null, read integer not null);
    create index notification_actorId on notifications (actorId);
    create index notification_subjectId on notifications (subjectId);
    create index notification_blockId on notifications (blockId);
    create index notification_read on notifications (read);
	create table nonces (value text primary key not null, address text not null, date integer not null);
    create table accounts (id text primary key not null, address text not null, created integer not null, lastSeen integer not null);
    create index account_address on accounts (address);
    create index account_lastSeen on accounts (lastSeen);
    create table account_threads (id text not null, accountId text not null, head text, skCipher blob not null, primary key (id, accountId));
    create index account_thread_accountId on account_threads (accountId);
    create table sessions (cafeId text primary key not null, access text not null, refresh text not null, expiry integer not null);
    create table cafe_requests (id text primary key not null, targetId text not null, cafeId text not null, type integer not null, date integer not null);
    create index cafe_request_cafeId on cafe_requests (cafeId);
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		return err
	}
	return nil
}
