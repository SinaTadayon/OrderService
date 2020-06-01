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
		Host:     config.Mongo.Host,
		Port:     config.Mongo.Port,
		Username: config.Mongo.User,
		//Password:     App.Cfg.Mongo.Pass,
		ConnTimeout:     time.Duration(config.Mongo.ConnectionTimeout) * time.Second,
		ReadTimeout:     time.Duration(config.Mongo.ReadTimeout) * time.Second,
		WriteTimeout:    time.Duration(config.Mongo.WriteTimeout) * time.Second,
		MaxConnIdleTime: time.Duration(config.Mongo.MaxConnIdleTime) * time.Second,
		MaxPoolSize:     uint64(config.Mongo.MaxPoolSize),
		MinPoolSize:     uint64(config.Mongo.MinPoolSize),
		WriteConcernW:   config.Mongo.WriteConcernW,
		WriteConcernJ:   config.Mongo.WriteConcernJ,
		RetryWrites:     config.Mongo.RetryWrite,
	}

	mongoAdapter, err = mongoadapter.NewMongo(mongoConf)
	if err != nil {
		applog.GLog.Logger.Error("mongoadapter.NewMongo failed", "error", err)
		os.Exit(1)
	}

	pkgItemRepo = pkg_repository.NewPkgItemRepository(mongoAdapter, config.Mongo.Database, config.Mongo.Collection)

	// Running Tests
	code := m.Run()
	// removeCollection()
	os.Exit(code)
}

func removeCollection() {
	if _, err := mongoAdapter.DeleteMany(config.Mongo.Database, config.Mongo.Collection, bson.M{}); err != nil {
	}
}

func TestExtractActionReason(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	pkgItem, _, _ := pkgItemRepo.FindPkgItmBuyinfById(ctx, uint64(2662550413), uint64(1000007))

	reason := ExtractActionReason(pkgItem.Subpackages[0].Tracking, "Cancel", Cancel, "Approval_Pending")
	fmt.Println(reason)
	assert.NotNil(t, reason)
}
