package reason

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/configs"
	pkg_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/pkg"
	applog "gitlab.faza.io/order-project/order-service/infrastructure/logger"
	"go.mongodb.org/mongo-driver/bson"
	"os"
	"testing"
	"time"
)

var pkgItemRepo pkg_repository.IPkgItemRepository
var mongoAdapter *mongoadapter.Mongo
var config *configs.Config

func TestMain(m *testing.M) {
	var path string
	var err error
	if os.Getenv("APP_MODE") == "dev" {
		path = "../../testdata/.env"
	} else {
		path = ""
	}

	applog.GLog.ZapLogger = applog.InitZap()
	applog.GLog.Logger = logger.NewZapLogger(applog.GLog.ZapLogger)

	config, _, err = configs.LoadConfigs(path, "")
	if err != nil {
		applog.GLog.Logger.Error("configs.LoadConfig failed",
			"error", err)
		os.Exit(1)
	}

	// store in mongo
	mongoConf := &mongoadapter.MongoConfig{
		//Host:     config.Mongo.Host,
		//Port:     config.Mongo.Port,
		ConnectUri: config.CmdMongo.Uri,
		Username:   config.CmdMongo.User,
		//Password:     App.Cfg.CmdMongo.Pass,
		ConnTimeout:            time.Duration(config.CmdMongo.ConnectionTimeout) * time.Second,
		ReadTimeout:            time.Duration(config.CmdMongo.ReadTimeout) * time.Second,
		WriteTimeout:           time.Duration(config.CmdMongo.WriteTimeout) * time.Second,
		MaxConnIdleTime:        time.Duration(config.CmdMongo.MaxConnIdleTime) * time.Second,
		HeartbeatInterval:      time.Duration(config.CmdMongo.HeartBeatInterval) * time.Second,
		ServerSelectionTimeout: time.Duration(config.CmdMongo.ServerSelectionTimeout) * time.Second,
		RetryConnect:           uint64(config.CmdMongo.RetryConnect),
		MaxPoolSize:            uint64(config.CmdMongo.MaxPoolSize),
		MinPoolSize:            uint64(config.CmdMongo.MinPoolSize),
		WriteConcernW:          config.CmdMongo.WriteConcernW,
		WriteConcernJ:          config.CmdMongo.WriteConcernJ,
		RetryWrites:            config.CmdMongo.RetryWrite,
	}

	mongoAdapter, err = mongoadapter.NewMongo(mongoConf)
	if err != nil {
		applog.GLog.Logger.Error("mongoadapter.NewMongo failed", "error", err)
		os.Exit(1)
	}

	pkgItemRepo = pkg_repository.NewPkgItemRepository(mongoAdapter, config.CmdMongo.Database, config.CmdMongo.Collection)

	// Running Tests
	code := m.Run()
	removeCollection()
	os.Exit(code)
}

func removeCollection() {
	if _, err := mongoAdapter.DeleteMany(config.CmdMongo.Database, config.CmdMongo.Collection, bson.M{}); err != nil {
	}
}

func TestExtractActionReason(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	pkgItem, _, _ := pkgItemRepo.FindPkgItmBuyinfById(ctx, uint64(2662550413), uint64(1000007))

	reason := ExtractActionReason(pkgItem.Subpackages[0].Tracking, "Cancel", Cancel, "Approval_Pending")
	fmt.Println(reason)
	assert.NotNil(t, reason)
}
