package cfgcommon

import (
	"context"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/infra/conf/geodata"
)

const confContextKey = configureLoadingContext(1)

type configureLoadingContext int

type configureLoadingEnvironment struct {
	geoLoader geodata.Loader
}

type ConfigureLoadingEnvironment interface {
	GetGeoLoader() geodata.Loader
}

func GetConfigureLoadingEnvironment(ctx context.Context) ConfigureLoadingEnvironment {
	return ctx.Value(confContextKey).(ConfigureLoadingEnvironment)
}

func NewConfigureLoadingContext(ctx context.Context) context.Context {
	environment := &configureLoadingEnvironment{}
	return context.WithValue(ctx, confContextKey, environment)
}

func SetGeoDataLoader(ctx context.Context, loader geodata.Loader) {
	GetConfigureLoadingEnvironment(ctx).(*configureLoadingEnvironment).geoLoader = loader
}

func (c *configureLoadingEnvironment) GetGeoLoader() geodata.Loader {
	if c.geoLoader == nil {
		var err error
		c.geoLoader, err = geodata.GetGeoDataLoader("standard")
		common.Must(err)
	}
	return c.geoLoader
}
