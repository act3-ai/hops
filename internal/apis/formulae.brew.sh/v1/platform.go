package v1

import (
	"fmt"

	jsonpatch "github.com/evanphx/json-patch/v5"
	easyjson "github.com/mailru/easyjson"

	"github.com/act3-ai/hops/internal/platform"
	"github.com/act3-ai/hops/internal/utils"
)

// ForPlatform produces the "compiled" metadata for the given platform by evaluating its variations.
func (info *Info) ForPlatform(plat platform.Platform) (*PlatformInfo, error) {
	base := info.PlatformInfo

	variation, ok := info.Variations[plat]
	if !ok || variation == nil {
		return &base, nil
	}

	platinfo, err := jsonPatch(&base, variation)
	if err != nil {
		return nil, fmt.Errorf("resolving %s metadata for platform %s: %w", info.Name, plat, err)
	}

	return platinfo, nil
}

func jsonPatch[T easyjson.MarshalerUnmarshaler](original, patch T) (T, error) {
	var newobj T

	ogjson, err := easyjson.Marshal(original)
	if err != nil {
		return newobj, fmt.Errorf("marshaling %T to JSON: %w", original, err)
	}

	patchjson, err := easyjson.Marshal(patch)
	if err != nil {
		return newobj, fmt.Errorf("marshaling %T to JSON: %w", patch, err)
	}

	// Apply the JSON merge patch
	newjson, err := jsonpatch.MergePatch(ogjson, patchjson)
	if err != nil {
		return newobj, fmt.Errorf("patching config: %w", err)
	}

	// Interpret empty configuration as nil
	if utils.BytesAreEmptyIsh(newjson) {
		// Returning nil is intentional.
		// Marshalling a non-nil object produces fields that were not present.
		return newobj, nil
	}

	newobj = *new(T)
	err = easyjson.Unmarshal(newjson, newobj)
	if err != nil {
		return newobj, fmt.Errorf("unmarshaling patched config into %T (%T): %w", newobj, newobj, err)
	}

	return newobj, nil
}
