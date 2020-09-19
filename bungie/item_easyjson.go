// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package bungie

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjsonA80d3b19DecodeGithubComRking788WarmindNetworkBungie(in *jlexer.Lexer, out *Record) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "Hash":
			out.Hash = uint(in.Uint())
		case "Name":
			out.Name = string(in.String())
		case "HasTitle":
			out.HasTitle = bool(in.Bool())
		case "MaleTitle":
			out.MaleTitle = string(in.String())
		case "FemaleTitle":
			out.FemaleTitle = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonA80d3b19EncodeGithubComRking788WarmindNetworkBungie(out *jwriter.Writer, in Record) {
	out.RawByte('{')
	first := true
	_ = first
	if in.Hash != 0 {
		const prefix string = ",\"Hash\":"
		first = false
		out.RawString(prefix[1:])
		out.Uint(uint(in.Hash))
	}
	if in.Name != "" {
		const prefix string = ",\"Name\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Name))
	}
	if in.HasTitle {
		const prefix string = ",\"HasTitle\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Bool(bool(in.HasTitle))
	}
	if in.MaleTitle != "" {
		const prefix string = ",\"MaleTitle\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.MaleTitle))
	}
	if in.FemaleTitle != "" {
		const prefix string = ",\"FemaleTitle\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.FemaleTitle))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Record) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonA80d3b19EncodeGithubComRking788WarmindNetworkBungie(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Record) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonA80d3b19EncodeGithubComRking788WarmindNetworkBungie(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Record) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonA80d3b19DecodeGithubComRking788WarmindNetworkBungie(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Record) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonA80d3b19DecodeGithubComRking788WarmindNetworkBungie(l, v)
}
func easyjsonA80d3b19DecodeGithubComRking788WarmindNetworkBungie1(in *jlexer.Lexer, out *ItemMetadata) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "TierType":
			out.TierType = int(in.Int())
		case "ClassType":
			out.ClassType = int(in.Int())
		case "BucketHash":
			out.BucketHash = uint(in.Uint())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonA80d3b19EncodeGithubComRking788WarmindNetworkBungie1(out *jwriter.Writer, in ItemMetadata) {
	out.RawByte('{')
	first := true
	_ = first
	if in.TierType != 0 {
		const prefix string = ",\"TierType\":"
		first = false
		out.RawString(prefix[1:])
		out.Int(int(in.TierType))
	}
	if in.ClassType != 0 {
		const prefix string = ",\"ClassType\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int(int(in.ClassType))
	}
	if in.BucketHash != 0 {
		const prefix string = ",\"BucketHash\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Uint(uint(in.BucketHash))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ItemMetadata) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonA80d3b19EncodeGithubComRking788WarmindNetworkBungie1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ItemMetadata) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonA80d3b19EncodeGithubComRking788WarmindNetworkBungie1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ItemMetadata) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonA80d3b19DecodeGithubComRking788WarmindNetworkBungie1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ItemMetadata) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonA80d3b19DecodeGithubComRking788WarmindNetworkBungie1(l, v)
}
func easyjsonA80d3b19DecodeGithubComRking788WarmindNetworkBungie2(in *jlexer.Lexer, out *ItemInstance) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "isEquipped":
			out.IsEquipped = bool(in.Bool())
		case "canEquip":
			out.CanEquip = bool(in.Bool())
		case "quality":
			out.Quality = int(in.Int())
		case "cannotEquipReason":
			out.CannotEquipReason = int(in.Int())
		case "DamageType":
			out.DamageType = int(in.Int())
		case "equipRequiredLevel":
			out.EquipRequiredLevel = int(in.Int())
		case "primaryStat":
			if in.IsNull() {
				in.Skip()
				out.PrimaryStat = nil
			} else {
				if out.PrimaryStat == nil {
					out.PrimaryStat = new(struct {
						StatHash     uint `json:"statHash"`
						Value        int  `json:"value"`
						MaximumValue int  `json:"maximumValue"`
						ItemLevel    int  `json:"itemLevel"`
					})
				}
				easyjsonA80d3b19Decode(in, out.PrimaryStat)
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonA80d3b19EncodeGithubComRking788WarmindNetworkBungie2(out *jwriter.Writer, in ItemInstance) {
	out.RawByte('{')
	first := true
	_ = first
	if in.IsEquipped {
		const prefix string = ",\"isEquipped\":"
		first = false
		out.RawString(prefix[1:])
		out.Bool(bool(in.IsEquipped))
	}
	if in.CanEquip {
		const prefix string = ",\"canEquip\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Bool(bool(in.CanEquip))
	}
	if in.Quality != 0 {
		const prefix string = ",\"quality\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int(int(in.Quality))
	}
	if in.CannotEquipReason != 0 {
		const prefix string = ",\"cannotEquipReason\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int(int(in.CannotEquipReason))
	}
	if in.DamageType != 0 {
		const prefix string = ",\"DamageType\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int(int(in.DamageType))
	}
	if in.EquipRequiredLevel != 0 {
		const prefix string = ",\"equipRequiredLevel\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int(int(in.EquipRequiredLevel))
	}
	if in.PrimaryStat != nil {
		const prefix string = ",\"primaryStat\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		easyjsonA80d3b19Encode(out, *in.PrimaryStat)
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ItemInstance) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonA80d3b19EncodeGithubComRking788WarmindNetworkBungie2(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ItemInstance) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonA80d3b19EncodeGithubComRking788WarmindNetworkBungie2(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ItemInstance) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonA80d3b19DecodeGithubComRking788WarmindNetworkBungie2(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ItemInstance) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonA80d3b19DecodeGithubComRking788WarmindNetworkBungie2(l, v)
}
func easyjsonA80d3b19Decode(in *jlexer.Lexer, out *struct {
	StatHash     uint `json:"statHash"`
	Value        int  `json:"value"`
	MaximumValue int  `json:"maximumValue"`
	ItemLevel    int  `json:"itemLevel"`
}) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "statHash":
			out.StatHash = uint(in.Uint())
		case "value":
			out.Value = int(in.Int())
		case "maximumValue":
			out.MaximumValue = int(in.Int())
		case "itemLevel":
			out.ItemLevel = int(in.Int())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonA80d3b19Encode(out *jwriter.Writer, in struct {
	StatHash     uint `json:"statHash"`
	Value        int  `json:"value"`
	MaximumValue int  `json:"maximumValue"`
	ItemLevel    int  `json:"itemLevel"`
}) {
	out.RawByte('{')
	first := true
	_ = first
	if in.StatHash != 0 {
		const prefix string = ",\"statHash\":"
		first = false
		out.RawString(prefix[1:])
		out.Uint(uint(in.StatHash))
	}
	if in.Value != 0 {
		const prefix string = ",\"value\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int(int(in.Value))
	}
	if in.MaximumValue != 0 {
		const prefix string = ",\"maximumValue\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int(int(in.MaximumValue))
	}
	if in.ItemLevel != 0 {
		const prefix string = ",\"itemLevel\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int(int(in.ItemLevel))
	}
	out.RawByte('}')
}
func easyjsonA80d3b19DecodeGithubComRking788WarmindNetworkBungie3(in *jlexer.Lexer, out *Item) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "itemHash":
			out.ItemHash = uint(in.Uint())
		case "itemInstanceId":
			out.InstanceID = string(in.String())
		case "bucketHash":
			out.BucketHash = uint(in.Uint())
		case "lockable":
			out.Lockable = bool(in.Bool())
		case "bindStatus":
			out.BindStatus = int(in.Int())
		case "state":
			out.State = int(in.Int())
		case "location":
			out.Location = int(in.Int())
		case "transferStatus":
			out.TransferStatus = int(in.Int())
		case "quantity":
			out.Quantity = int(in.Int())
		case "Instance":
			if in.IsNull() {
				in.Skip()
				out.Instance = nil
			} else {
				if out.Instance == nil {
					out.Instance = new(ItemInstance)
				}
				(*out.Instance).UnmarshalEasyJSON(in)
			}
		case "Character":
			if in.IsNull() {
				in.Skip()
				out.Character = nil
			} else {
				if out.Character == nil {
					out.Character = new(Character)
				}
				(*out.Character).UnmarshalEasyJSON(in)
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonA80d3b19EncodeGithubComRking788WarmindNetworkBungie3(out *jwriter.Writer, in Item) {
	out.RawByte('{')
	first := true
	_ = first
	if in.ItemHash != 0 {
		const prefix string = ",\"itemHash\":"
		first = false
		out.RawString(prefix[1:])
		out.Uint(uint(in.ItemHash))
	}
	if in.InstanceID != "" {
		const prefix string = ",\"itemInstanceId\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.InstanceID))
	}
	if in.BucketHash != 0 {
		const prefix string = ",\"bucketHash\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Uint(uint(in.BucketHash))
	}
	if in.Lockable {
		const prefix string = ",\"lockable\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Bool(bool(in.Lockable))
	}
	if in.BindStatus != 0 {
		const prefix string = ",\"bindStatus\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int(int(in.BindStatus))
	}
	if in.State != 0 {
		const prefix string = ",\"state\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int(int(in.State))
	}
	if in.Location != 0 {
		const prefix string = ",\"location\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int(int(in.Location))
	}
	if in.TransferStatus != 0 {
		const prefix string = ",\"transferStatus\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int(int(in.TransferStatus))
	}
	if in.Quantity != 0 {
		const prefix string = ",\"quantity\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int(int(in.Quantity))
	}
	if in.Instance != nil {
		const prefix string = ",\"Instance\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		(*in.Instance).MarshalEasyJSON(out)
	}
	if in.Character != nil {
		const prefix string = ",\"Character\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		(*in.Character).MarshalEasyJSON(out)
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Item) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonA80d3b19EncodeGithubComRking788WarmindNetworkBungie3(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Item) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonA80d3b19EncodeGithubComRking788WarmindNetworkBungie3(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Item) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonA80d3b19DecodeGithubComRking788WarmindNetworkBungie3(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Item) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonA80d3b19DecodeGithubComRking788WarmindNetworkBungie3(l, v)
}