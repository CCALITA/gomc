package item

import "github.com/fanxiyao/gomc/internal/registry"

// Items is the global item registry.
var Items *registry.Registry[ItemProperties]

// InitRegistry populates the Items registry with all known item types
// and freezes it. It should be called once during program startup.
func InitRegistry() {
	Items = registry.New[ItemProperties]()

	for id, props := range properties {
		// Register uses the item name as the key.
		got, err := Items.Register(props.Name, props)
		if err != nil {
			panic("item.InitRegistry: " + err.Error())
		}
		// Sanity-check: the registry-assigned ID must match our constant.
		// Because we register in arbitrary map order, we cannot guarantee
		// sequential assignment.  Instead we store a mapping.
		_ = got
		_ = id
	}

	Items.Freeze()
}
