package sdkutil

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func RenderQueryResponse(cdc *codec.Codec, obj interface{}) ([]byte, *sdkerrors.Error) {
	response, err := codec.MarshalJSONIndent(cdc, obj)
	if err != nil {
		return nil, sdkerrors.New("sdkutil", 1, err.Error())
	}
	return response, nil
}
