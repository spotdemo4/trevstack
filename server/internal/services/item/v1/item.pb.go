// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.5
// 	protoc        (unknown)
// source: item/v1/item.proto

package itemv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Item struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            *int64                 `protobuf:"varint,1,opt,name=id,proto3,oneof" json:"id,omitempty"`
	Name          string                 `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Description   string                 `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	Price         float32                `protobuf:"fixed32,4,opt,name=price,proto3" json:"price,omitempty"`
	Quantity      int32                  `protobuf:"varint,5,opt,name=quantity,proto3" json:"quantity,omitempty"`
	Added         *timestamppb.Timestamp `protobuf:"bytes,6,opt,name=added,proto3,oneof" json:"added,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Item) Reset() {
	*x = Item{}
	mi := &file_item_v1_item_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Item) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Item) ProtoMessage() {}

func (x *Item) ProtoReflect() protoreflect.Message {
	mi := &file_item_v1_item_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Item.ProtoReflect.Descriptor instead.
func (*Item) Descriptor() ([]byte, []int) {
	return file_item_v1_item_proto_rawDescGZIP(), []int{0}
}

func (x *Item) GetId() int64 {
	if x != nil && x.Id != nil {
		return *x.Id
	}
	return 0
}

func (x *Item) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Item) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Item) GetPrice() float32 {
	if x != nil {
		return x.Price
	}
	return 0
}

func (x *Item) GetQuantity() int32 {
	if x != nil {
		return x.Quantity
	}
	return 0
}

func (x *Item) GetAdded() *timestamppb.Timestamp {
	if x != nil {
		return x.Added
	}
	return nil
}

type GetItemRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            int64                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetItemRequest) Reset() {
	*x = GetItemRequest{}
	mi := &file_item_v1_item_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetItemRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetItemRequest) ProtoMessage() {}

func (x *GetItemRequest) ProtoReflect() protoreflect.Message {
	mi := &file_item_v1_item_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetItemRequest.ProtoReflect.Descriptor instead.
func (*GetItemRequest) Descriptor() ([]byte, []int) {
	return file_item_v1_item_proto_rawDescGZIP(), []int{1}
}

func (x *GetItemRequest) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

type GetItemResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Item          *Item                  `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetItemResponse) Reset() {
	*x = GetItemResponse{}
	mi := &file_item_v1_item_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetItemResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetItemResponse) ProtoMessage() {}

func (x *GetItemResponse) ProtoReflect() protoreflect.Message {
	mi := &file_item_v1_item_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetItemResponse.ProtoReflect.Descriptor instead.
func (*GetItemResponse) Descriptor() ([]byte, []int) {
	return file_item_v1_item_proto_rawDescGZIP(), []int{2}
}

func (x *GetItemResponse) GetItem() *Item {
	if x != nil {
		return x.Item
	}
	return nil
}

type GetItemsRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Start         *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=start,proto3,oneof" json:"start,omitempty"`
	End           *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=end,proto3,oneof" json:"end,omitempty"`
	Filter        *string                `protobuf:"bytes,3,opt,name=filter,proto3,oneof" json:"filter,omitempty"`
	Limit         *int32                 `protobuf:"varint,4,opt,name=limit,proto3,oneof" json:"limit,omitempty"`
	Offset        *int32                 `protobuf:"varint,5,opt,name=offset,proto3,oneof" json:"offset,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetItemsRequest) Reset() {
	*x = GetItemsRequest{}
	mi := &file_item_v1_item_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetItemsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetItemsRequest) ProtoMessage() {}

func (x *GetItemsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_item_v1_item_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetItemsRequest.ProtoReflect.Descriptor instead.
func (*GetItemsRequest) Descriptor() ([]byte, []int) {
	return file_item_v1_item_proto_rawDescGZIP(), []int{3}
}

func (x *GetItemsRequest) GetStart() *timestamppb.Timestamp {
	if x != nil {
		return x.Start
	}
	return nil
}

func (x *GetItemsRequest) GetEnd() *timestamppb.Timestamp {
	if x != nil {
		return x.End
	}
	return nil
}

func (x *GetItemsRequest) GetFilter() string {
	if x != nil && x.Filter != nil {
		return *x.Filter
	}
	return ""
}

func (x *GetItemsRequest) GetLimit() int32 {
	if x != nil && x.Limit != nil {
		return *x.Limit
	}
	return 0
}

func (x *GetItemsRequest) GetOffset() int32 {
	if x != nil && x.Offset != nil {
		return *x.Offset
	}
	return 0
}

type GetItemsResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Items         []*Item                `protobuf:"bytes,1,rep,name=items,proto3" json:"items,omitempty"`
	Count         int64                  `protobuf:"varint,2,opt,name=count,proto3" json:"count,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetItemsResponse) Reset() {
	*x = GetItemsResponse{}
	mi := &file_item_v1_item_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetItemsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetItemsResponse) ProtoMessage() {}

func (x *GetItemsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_item_v1_item_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetItemsResponse.ProtoReflect.Descriptor instead.
func (*GetItemsResponse) Descriptor() ([]byte, []int) {
	return file_item_v1_item_proto_rawDescGZIP(), []int{4}
}

func (x *GetItemsResponse) GetItems() []*Item {
	if x != nil {
		return x.Items
	}
	return nil
}

func (x *GetItemsResponse) GetCount() int64 {
	if x != nil {
		return x.Count
	}
	return 0
}

type CreateItemRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Item          *Item                  `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateItemRequest) Reset() {
	*x = CreateItemRequest{}
	mi := &file_item_v1_item_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateItemRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateItemRequest) ProtoMessage() {}

func (x *CreateItemRequest) ProtoReflect() protoreflect.Message {
	mi := &file_item_v1_item_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateItemRequest.ProtoReflect.Descriptor instead.
func (*CreateItemRequest) Descriptor() ([]byte, []int) {
	return file_item_v1_item_proto_rawDescGZIP(), []int{5}
}

func (x *CreateItemRequest) GetItem() *Item {
	if x != nil {
		return x.Item
	}
	return nil
}

type CreateItemResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Item          *Item                  `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateItemResponse) Reset() {
	*x = CreateItemResponse{}
	mi := &file_item_v1_item_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateItemResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateItemResponse) ProtoMessage() {}

func (x *CreateItemResponse) ProtoReflect() protoreflect.Message {
	mi := &file_item_v1_item_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateItemResponse.ProtoReflect.Descriptor instead.
func (*CreateItemResponse) Descriptor() ([]byte, []int) {
	return file_item_v1_item_proto_rawDescGZIP(), []int{6}
}

func (x *CreateItemResponse) GetItem() *Item {
	if x != nil {
		return x.Item
	}
	return nil
}

type UpdateItemRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Item          *Item                  `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateItemRequest) Reset() {
	*x = UpdateItemRequest{}
	mi := &file_item_v1_item_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateItemRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateItemRequest) ProtoMessage() {}

func (x *UpdateItemRequest) ProtoReflect() protoreflect.Message {
	mi := &file_item_v1_item_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateItemRequest.ProtoReflect.Descriptor instead.
func (*UpdateItemRequest) Descriptor() ([]byte, []int) {
	return file_item_v1_item_proto_rawDescGZIP(), []int{7}
}

func (x *UpdateItemRequest) GetItem() *Item {
	if x != nil {
		return x.Item
	}
	return nil
}

type UpdateItemResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Item          *Item                  `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateItemResponse) Reset() {
	*x = UpdateItemResponse{}
	mi := &file_item_v1_item_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateItemResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateItemResponse) ProtoMessage() {}

func (x *UpdateItemResponse) ProtoReflect() protoreflect.Message {
	mi := &file_item_v1_item_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateItemResponse.ProtoReflect.Descriptor instead.
func (*UpdateItemResponse) Descriptor() ([]byte, []int) {
	return file_item_v1_item_proto_rawDescGZIP(), []int{8}
}

func (x *UpdateItemResponse) GetItem() *Item {
	if x != nil {
		return x.Item
	}
	return nil
}

type DeleteItemRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            int64                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DeleteItemRequest) Reset() {
	*x = DeleteItemRequest{}
	mi := &file_item_v1_item_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DeleteItemRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteItemRequest) ProtoMessage() {}

func (x *DeleteItemRequest) ProtoReflect() protoreflect.Message {
	mi := &file_item_v1_item_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteItemRequest.ProtoReflect.Descriptor instead.
func (*DeleteItemRequest) Descriptor() ([]byte, []int) {
	return file_item_v1_item_proto_rawDescGZIP(), []int{9}
}

func (x *DeleteItemRequest) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

type DeleteItemResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DeleteItemResponse) Reset() {
	*x = DeleteItemResponse{}
	mi := &file_item_v1_item_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DeleteItemResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteItemResponse) ProtoMessage() {}

func (x *DeleteItemResponse) ProtoReflect() protoreflect.Message {
	mi := &file_item_v1_item_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteItemResponse.ProtoReflect.Descriptor instead.
func (*DeleteItemResponse) Descriptor() ([]byte, []int) {
	return file_item_v1_item_proto_rawDescGZIP(), []int{10}
}

var File_item_v1_item_proto protoreflect.FileDescriptor

var file_item_v1_item_proto_rawDesc = string([]byte{
	0x0a, 0x12, 0x69, 0x74, 0x65, 0x6d, 0x2f, 0x76, 0x31, 0x2f, 0x69, 0x74, 0x65, 0x6d, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x69, 0x74, 0x65, 0x6d, 0x2e, 0x76, 0x31, 0x1a, 0x1f, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xcb,
	0x01, 0x0a, 0x04, 0x49, 0x74, 0x65, 0x6d, 0x12, 0x13, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x03, 0x48, 0x00, 0x52, 0x02, 0x69, 0x64, 0x88, 0x01, 0x01, 0x12, 0x12, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x12, 0x14, 0x0a, 0x05, 0x70, 0x72, 0x69, 0x63, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x02, 0x52, 0x05, 0x70, 0x72, 0x69, 0x63, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x71, 0x75, 0x61, 0x6e,
	0x74, 0x69, 0x74, 0x79, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x71, 0x75, 0x61, 0x6e,
	0x74, 0x69, 0x74, 0x79, 0x12, 0x35, 0x0a, 0x05, 0x61, 0x64, 0x64, 0x65, 0x64, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x48,
	0x01, 0x52, 0x05, 0x61, 0x64, 0x64, 0x65, 0x64, 0x88, 0x01, 0x01, 0x42, 0x05, 0x0a, 0x03, 0x5f,
	0x69, 0x64, 0x42, 0x08, 0x0a, 0x06, 0x5f, 0x61, 0x64, 0x64, 0x65, 0x64, 0x22, 0x20, 0x0a, 0x0e,
	0x47, 0x65, 0x74, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x22, 0x34,
	0x0a, 0x0f, 0x47, 0x65, 0x74, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x21, 0x0a, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x0d, 0x2e, 0x69, 0x74, 0x65, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x04,
	0x69, 0x74, 0x65, 0x6d, 0x22, 0x82, 0x02, 0x0a, 0x0f, 0x47, 0x65, 0x74, 0x49, 0x74, 0x65, 0x6d,
	0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x35, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x72,
	0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x48, 0x00, 0x52, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x88, 0x01, 0x01, 0x12,
	0x31, 0x0a, 0x03, 0x65, 0x6e, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x48, 0x01, 0x52, 0x03, 0x65, 0x6e, 0x64, 0x88,
	0x01, 0x01, 0x12, 0x1b, 0x0a, 0x06, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x48, 0x02, 0x52, 0x06, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x88, 0x01, 0x01, 0x12,
	0x19, 0x0a, 0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x48, 0x03,
	0x52, 0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x88, 0x01, 0x01, 0x12, 0x1b, 0x0a, 0x06, 0x6f, 0x66,
	0x66, 0x73, 0x65, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x48, 0x04, 0x52, 0x06, 0x6f, 0x66,
	0x66, 0x73, 0x65, 0x74, 0x88, 0x01, 0x01, 0x42, 0x08, 0x0a, 0x06, 0x5f, 0x73, 0x74, 0x61, 0x72,
	0x74, 0x42, 0x06, 0x0a, 0x04, 0x5f, 0x65, 0x6e, 0x64, 0x42, 0x09, 0x0a, 0x07, 0x5f, 0x66, 0x69,
	0x6c, 0x74, 0x65, 0x72, 0x42, 0x08, 0x0a, 0x06, 0x5f, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x42, 0x09,
	0x0a, 0x07, 0x5f, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x22, 0x4d, 0x0a, 0x10, 0x47, 0x65, 0x74,
	0x49, 0x74, 0x65, 0x6d, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x23, 0x0a,
	0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x69,
	0x74, 0x65, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x05, 0x69, 0x74, 0x65,
	0x6d, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0x36, 0x0a, 0x11, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x21, 0x0a,
	0x04, 0x69, 0x74, 0x65, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x69, 0x74,
	0x65, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x04, 0x69, 0x74, 0x65, 0x6d,
	0x22, 0x37, 0x0a, 0x12, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x21, 0x0a, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x69, 0x74, 0x65, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x49,
	0x74, 0x65, 0x6d, 0x52, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x22, 0x36, 0x0a, 0x11, 0x55, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x21,
	0x0a, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x69,
	0x74, 0x65, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x04, 0x69, 0x74, 0x65,
	0x6d, 0x22, 0x37, 0x0a, 0x12, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x21, 0x0a, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x69, 0x74, 0x65, 0x6d, 0x2e, 0x76, 0x31, 0x2e,
	0x49, 0x74, 0x65, 0x6d, 0x52, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x22, 0x23, 0x0a, 0x11, 0x44, 0x65,
	0x6c, 0x65, 0x74, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x22,
	0x14, 0x0a, 0x12, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0xeb, 0x02, 0x0a, 0x0b, 0x49, 0x74, 0x65, 0x6d, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x3e, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x49, 0x74, 0x65, 0x6d,
	0x12, 0x17, 0x2e, 0x69, 0x74, 0x65, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x49, 0x74,
	0x65, 0x6d, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x69, 0x74, 0x65, 0x6d,
	0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x41, 0x0a, 0x08, 0x47, 0x65, 0x74, 0x49, 0x74, 0x65, 0x6d,
	0x73, 0x12, 0x18, 0x2e, 0x69, 0x74, 0x65, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x49,
	0x74, 0x65, 0x6d, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x69, 0x74,
	0x65, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x49, 0x74, 0x65, 0x6d, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x47, 0x0a, 0x0a, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x12, 0x1a, 0x2e, 0x69, 0x74, 0x65, 0x6d, 0x2e, 0x76, 0x31,
	0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x69, 0x74, 0x65, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x00, 0x12, 0x47, 0x0a, 0x0a, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x12,
	0x1a, 0x2e, 0x69, 0x74, 0x65, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x49, 0x74, 0x65, 0x6d, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x69, 0x74,
	0x65, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x49, 0x74, 0x65, 0x6d,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x47, 0x0a, 0x0a, 0x44, 0x65,
	0x6c, 0x65, 0x74, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x12, 0x1a, 0x2e, 0x69, 0x74, 0x65, 0x6d, 0x2e,
	0x76, 0x31, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x69, 0x74, 0x65, 0x6d, 0x2e, 0x76, 0x31, 0x2e, 0x44,
	0x65, 0x6c, 0x65, 0x74, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x00, 0x42, 0x9d, 0x01, 0x0a, 0x0b, 0x63, 0x6f, 0x6d, 0x2e, 0x69, 0x74, 0x65, 0x6d,
	0x2e, 0x76, 0x31, 0x42, 0x09, 0x49, 0x74, 0x65, 0x6d, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01,
	0x5a, 0x46, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x70, 0x6f,
	0x74, 0x64, 0x65, 0x6d, 0x6f, 0x34, 0x2f, 0x74, 0x72, 0x65, 0x76, 0x73, 0x74, 0x61, 0x63, 0x6b,
	0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c,
	0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x2f, 0x69, 0x74, 0x65, 0x6d, 0x2f, 0x76,
	0x31, 0x3b, 0x69, 0x74, 0x65, 0x6d, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x49, 0x58, 0x58, 0xaa, 0x02,
	0x07, 0x49, 0x74, 0x65, 0x6d, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x07, 0x49, 0x74, 0x65, 0x6d, 0x5c,
	0x56, 0x31, 0xe2, 0x02, 0x13, 0x49, 0x74, 0x65, 0x6d, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42,
	0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x08, 0x49, 0x74, 0x65, 0x6d, 0x3a,
	0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_item_v1_item_proto_rawDescOnce sync.Once
	file_item_v1_item_proto_rawDescData []byte
)

func file_item_v1_item_proto_rawDescGZIP() []byte {
	file_item_v1_item_proto_rawDescOnce.Do(func() {
		file_item_v1_item_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_item_v1_item_proto_rawDesc), len(file_item_v1_item_proto_rawDesc)))
	})
	return file_item_v1_item_proto_rawDescData
}

var file_item_v1_item_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_item_v1_item_proto_goTypes = []any{
	(*Item)(nil),                  // 0: item.v1.Item
	(*GetItemRequest)(nil),        // 1: item.v1.GetItemRequest
	(*GetItemResponse)(nil),       // 2: item.v1.GetItemResponse
	(*GetItemsRequest)(nil),       // 3: item.v1.GetItemsRequest
	(*GetItemsResponse)(nil),      // 4: item.v1.GetItemsResponse
	(*CreateItemRequest)(nil),     // 5: item.v1.CreateItemRequest
	(*CreateItemResponse)(nil),    // 6: item.v1.CreateItemResponse
	(*UpdateItemRequest)(nil),     // 7: item.v1.UpdateItemRequest
	(*UpdateItemResponse)(nil),    // 8: item.v1.UpdateItemResponse
	(*DeleteItemRequest)(nil),     // 9: item.v1.DeleteItemRequest
	(*DeleteItemResponse)(nil),    // 10: item.v1.DeleteItemResponse
	(*timestamppb.Timestamp)(nil), // 11: google.protobuf.Timestamp
}
var file_item_v1_item_proto_depIdxs = []int32{
	11, // 0: item.v1.Item.added:type_name -> google.protobuf.Timestamp
	0,  // 1: item.v1.GetItemResponse.item:type_name -> item.v1.Item
	11, // 2: item.v1.GetItemsRequest.start:type_name -> google.protobuf.Timestamp
	11, // 3: item.v1.GetItemsRequest.end:type_name -> google.protobuf.Timestamp
	0,  // 4: item.v1.GetItemsResponse.items:type_name -> item.v1.Item
	0,  // 5: item.v1.CreateItemRequest.item:type_name -> item.v1.Item
	0,  // 6: item.v1.CreateItemResponse.item:type_name -> item.v1.Item
	0,  // 7: item.v1.UpdateItemRequest.item:type_name -> item.v1.Item
	0,  // 8: item.v1.UpdateItemResponse.item:type_name -> item.v1.Item
	1,  // 9: item.v1.ItemService.GetItem:input_type -> item.v1.GetItemRequest
	3,  // 10: item.v1.ItemService.GetItems:input_type -> item.v1.GetItemsRequest
	5,  // 11: item.v1.ItemService.CreateItem:input_type -> item.v1.CreateItemRequest
	7,  // 12: item.v1.ItemService.UpdateItem:input_type -> item.v1.UpdateItemRequest
	9,  // 13: item.v1.ItemService.DeleteItem:input_type -> item.v1.DeleteItemRequest
	2,  // 14: item.v1.ItemService.GetItem:output_type -> item.v1.GetItemResponse
	4,  // 15: item.v1.ItemService.GetItems:output_type -> item.v1.GetItemsResponse
	6,  // 16: item.v1.ItemService.CreateItem:output_type -> item.v1.CreateItemResponse
	8,  // 17: item.v1.ItemService.UpdateItem:output_type -> item.v1.UpdateItemResponse
	10, // 18: item.v1.ItemService.DeleteItem:output_type -> item.v1.DeleteItemResponse
	14, // [14:19] is the sub-list for method output_type
	9,  // [9:14] is the sub-list for method input_type
	9,  // [9:9] is the sub-list for extension type_name
	9,  // [9:9] is the sub-list for extension extendee
	0,  // [0:9] is the sub-list for field type_name
}

func init() { file_item_v1_item_proto_init() }
func file_item_v1_item_proto_init() {
	if File_item_v1_item_proto != nil {
		return
	}
	file_item_v1_item_proto_msgTypes[0].OneofWrappers = []any{}
	file_item_v1_item_proto_msgTypes[3].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_item_v1_item_proto_rawDesc), len(file_item_v1_item_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_item_v1_item_proto_goTypes,
		DependencyIndexes: file_item_v1_item_proto_depIdxs,
		MessageInfos:      file_item_v1_item_proto_msgTypes,
	}.Build()
	File_item_v1_item_proto = out.File
	file_item_v1_item_proto_goTypes = nil
	file_item_v1_item_proto_depIdxs = nil
}
