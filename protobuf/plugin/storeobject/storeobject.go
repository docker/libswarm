package storeobject

import (
	"github.com/docker/swarmkit/protobuf/plugin"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
)

// FIXME(aaronl): Look at fields inside the descriptor instead of
// special-casing based on name.
var typesWithNoSpec = map[string]struct{}{
	"Task":      {},
	"Resource":  {},
	"Extension": {},
}

type storeObjectGen struct {
	*generator.Generator
	generator.PluginImports
	eventsPkg  generator.Single
	stringsPkg generator.Single
}

func init() {
	generator.RegisterPlugin(new(storeObjectGen))
}

func (d *storeObjectGen) Name() string {
	return "storeobject"
}

func (d *storeObjectGen) Init(g *generator.Generator) {
	d.Generator = g
}

func (d *storeObjectGen) genMsgStoreObject(m *generator.Descriptor, storeObject *plugin.StoreObject) {
	ccTypeName := generator.CamelCaseSlice(m.TypeName())

	// Generate event types

	d.P("type ", ccTypeName, "CheckFunc func(t1, t2 *", ccTypeName, ") bool")
	d.P()

	for _, event := range []string{"Create", "Update", "Delete"} {
		d.P("type Event", event, ccTypeName, " struct {")
		d.In()
		d.P(ccTypeName, " *", ccTypeName)
		d.P("Checks []", ccTypeName, "CheckFunc")
		d.Out()
		d.P("}")
		d.P()
		d.P("func (e Event", event, ccTypeName, ") Matches(apiEvent ", d.eventsPkg.Use(), ".Event) bool {")
		d.In()
		d.P("typedEvent, ok := apiEvent.(Event", event, ccTypeName, ")")
		d.P("if !ok {")
		d.In()
		d.P("return false")
		d.Out()
		d.P("}")
		d.P()
		d.P("for _, check := range e.Checks {")
		d.In()
		d.P("if !check(e.", ccTypeName, ", typedEvent.", ccTypeName, ") {")
		d.In()
		d.P("return false")
		d.Out()
		d.P("}")
		d.Out()
		d.P("}")
		d.P("return true")
		d.Out()
		d.P("}")
	}

	// Generate methods for this type

	d.P("func (m *", ccTypeName, ") CopyStoreObject() StoreObject {")
	d.In()
	d.P("return m.Copy()")
	d.Out()
	d.P("}")
	d.P()

	d.P("func (m *", ccTypeName, ") GetMeta() Meta {")
	d.In()
	d.P("return m.Meta")
	d.Out()
	d.P("}")
	d.P()

	d.P("func (m *", ccTypeName, ") SetMeta(meta Meta) {")
	d.In()
	d.P("m.Meta = meta")
	d.Out()
	d.P("}")
	d.P()

	d.P("func (m *", ccTypeName, ") GetID() string {")
	d.In()
	d.P("return m.ID")
	d.Out()
	d.P("}")
	d.P()

	d.P("func (m *", ccTypeName, ") EventCreate() Event {")
	d.In()
	d.P("return EventCreate", ccTypeName, "{", ccTypeName, ": m}")
	d.Out()
	d.P("}")
	d.P()

	d.P("func (m *", ccTypeName, ") EventUpdate() Event {")
	d.In()
	d.P("return EventUpdate", ccTypeName, "{", ccTypeName, ": m}")
	d.Out()
	d.P("}")
	d.P()

	d.P("func (m *", ccTypeName, ") EventDelete() Event {")
	d.In()
	d.P("return EventDelete", ccTypeName, "{", ccTypeName, ": m}")
	d.Out()
	d.P("}")
	d.P()

	// Generate indexer by ID

	d.P("type ", ccTypeName, "IndexerByID struct{}")
	d.P()

	d.genFromArgs(ccTypeName + "IndexerByID")
	d.genPrefixFromArgs(ccTypeName + "IndexerByID")

	d.P("func (indexer ", ccTypeName, "IndexerByID) FromObject(obj interface{}) (bool, []byte, error) {")
	d.In()
	d.P("m := obj.(*", ccTypeName, ")")
	// Add the null character as a terminator
	d.P(`return true, []byte(m.ID + "\x00"), nil`)
	d.Out()
	d.P("}")

	// Generate indexer by name

	d.P("type ", ccTypeName, "IndexerByName struct{}")
	d.P()

	d.genFromArgs(ccTypeName + "IndexerByName")
	d.genPrefixFromArgs(ccTypeName + "IndexerByName")

	d.P("func (indexer ", ccTypeName, "IndexerByName) FromObject(obj interface{}) (bool, []byte, error) {")
	d.In()
	d.P("m := obj.(*", ccTypeName, ")")
	if _, hasNoSpec := typesWithNoSpec[*m.Name]; hasNoSpec {
		d.P(`val := m.Annotations.Name`)
	} else {
		d.P(`val := m.Spec.Annotations.Name`)
	}
	// Add the null character as a terminator
	d.P("return true, []byte(", d.stringsPkg.Use(), `.ToLower(val) + "\x00"), nil`)
	d.Out()
	d.P("}")

	// Generate custom indexer

	d.P("type ", ccTypeName, "CustomIndexer struct{}")
	d.P()

	d.genFromArgs(ccTypeName + "CustomIndexer")
	d.genPrefixFromArgs(ccTypeName + "CustomIndexer")

	d.P("func (indexer ", ccTypeName, "CustomIndexer) FromObject(obj interface{}) (bool, [][]byte, error) {")
	d.In()
	d.P("m := obj.(*", ccTypeName, ")")
	if _, hasNoSpec := typesWithNoSpec[*m.Name]; hasNoSpec {
		d.P(`return customIndexer("", &m.Annotations)`)
	} else {
		d.P(`return customIndexer("", &m.Spec.Annotations)`)
	}
	d.Out()
	d.P("}")
}

func (d *storeObjectGen) genFromArgs(indexerName string) {
	d.P("func (indexer ", indexerName, ") FromArgs(args ...interface{}) ([]byte, error) {")
	d.In()
	d.P("return fromArgs(args...)")
	d.Out()
	d.P("}")
}

func (d *storeObjectGen) genPrefixFromArgs(indexerName string) {
	d.P("func (indexer ", indexerName, ") PrefixFromArgs(args ...interface{}) ([]byte, error) {")
	d.In()
	d.P("return prefixFromArgs(args...)")
	d.Out()
	d.P("}")

}

func (d *storeObjectGen) Generate(file *generator.FileDescriptor) {
	d.PluginImports = generator.NewPluginImports(d.Generator)
	d.eventsPkg = d.NewImport("github.com/docker/go-events")
	d.stringsPkg = d.NewImport("strings")

	for _, m := range file.Messages() {
		if m.DescriptorProto.GetOptions().GetMapEntry() {
			continue
		}

		if m.Options == nil {
			continue
		}
		storeObjIntf, err := proto.GetExtension(m.Options, plugin.E_StoreObject)
		if err != nil {
			// no StoreObject extension
			continue
		}

		d.genMsgStoreObject(m, storeObjIntf.(*plugin.StoreObject))
	}
}
