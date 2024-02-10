package maps

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (path MapPath) String() string {
	return string(path)
}

func (path MapPath) ParsePath() MapPathPieces {
	inQuotes := false
	pieces := []string{}
	var sb strings.Builder
	for _, char := range path {
		if char == '.' {
			if inQuotes {
				sb.WriteRune(char)
			} else {
				pieces = append(pieces, sb.String())
				sb.Reset()
			}
		} else {
			if char == '\'' || char == '"' {
				inQuotes = !inQuotes
			} else {
				sb.WriteRune(char)
			}
		}
	}
	if sb.Len() > 0 {
		pieces = append(pieces, sb.String())
		sb.Reset()
	}
	return pieces
}

func (pieces MapPathPieces) ToMapPath() MapPath {
	var sb strings.Builder
	for i, piece := range pieces {
		pieceString := piece
		if strings.Contains(piece, ".") {
			pieceString = fmt.Sprintf("'%s'", piece)
		}
		if i == 0 {
			sb.WriteString(pieceString)
		} else {
			sb.WriteString("." + pieceString)
		}
	}
	return MapPath(sb.String())
}

func (pieces MapPathPieces) RemoveFirstPiece() MapPath {
	return MapPathPieces(pieces[1:]).ToMapPath()
}

func NewMapPath(path string) MapPath {
	return MapPath(path)
}

func GetStringValueFromMapByPath(path MapPath, tknObj map[string]interface{}, throwIfNotString bool) (string, error) {
	value, err := GetValueFromMapByPath(path, tknObj)
	if err != nil {
		return "", fmt.Errorf("path: %s - %v", path, err)
	}
	switch t := value.(type) {
	case string:
		return t, nil
	case int, int32, int64:
		return fmt.Sprintf("%v", t), nil
	case float32:
		var i int = int(t)
		return fmt.Sprintf("%v", i), nil
	case float64:
		var i int = int(t)
		return fmt.Sprintf("%v", i), nil
	default:

		if throwIfNotString {
			return "", fmt.Errorf("the value is of type '%T' and not a string", value)
		} else {
			if t == nil {
				return "", nil
			}
			return fmt.Sprintf("%v", t), nil
		}
	}
}

func GetValueFromMapByPath(path MapPath, tknObj map[string]interface{}) (interface{}, error) {
	pathSections := path.ParsePath()
	section, exists := tknObj[pathSections[0]]
	if !exists {
		return nil, fmt.Errorf("section '%s' does not exist", pathSections[0])
	}
	if len(pathSections) == 1 {
		return section, nil
	}
	switch t := section.(type) {
	case map[string]interface{}:
		return GetValueFromMapByPath(pathSections.RemoveFirstPiece(), t)
	case map[string]string:
		newSectionData := map[string]interface{}{}
		for k, v := range t {
			newSectionData[k] = v
		}
		return GetValueFromMapByPath(pathSections.RemoveFirstPiece(), newSectionData)
	default:
		return nil, fmt.Errorf("section '%s' is of type %T...cannot continue", pathSections[0], section)
	}
}

func DeepCopyMap(mapObj map[string]interface{}) (map[string]interface{}, error) {
	newMap := make(map[string]interface{})
	newMapBytes, err := json.Marshal(mapObj)
	if err != nil {
		return newMap, err
	}
	err = json.Unmarshal(newMapBytes, &newMap)
	if err != nil {
		return newMap, err
	}
	return newMap, nil

}

func DeepCopyMapPointer(mapObj *map[string]interface{}) (*map[string]interface{}, error) {
	newMap := make(map[string]interface{})
	newMapBytes, err := json.Marshal(&mapObj)
	if err != nil {
		return &newMap, err
	}
	err = json.Unmarshal(newMapBytes, &newMap)
	if err != nil {
		return &newMap, err
	}
	return &newMap, nil

}

func DeepCopyInterface(mapObj interface{}) (interface{}, error) {
	var newInt interface{}
	newIntBytes, err := json.Marshal(mapObj)
	if err != nil {
		return newInt, err
	}
	err = json.Unmarshal(newIntBytes, &newInt)
	if err != nil {
		return newInt, err
	}
	return newInt, nil

}

// DeepCopy will create a deep copy of this map. The depth of this
// copy is all inclusive. Both maps and slices will be considered when
// making the copy.
func (m CopyableMap) DeepCopy() map[string]interface{} {
	result := map[string]interface{}{}

	for k, v := range m {
		// Handle maps
		mapvalue, isMap := v.(map[string]interface{})
		if isMap {
			result[k] = CopyableMap(mapvalue).DeepCopy()
			continue
		}

		// Handle slices
		slicevalue, isSlice := v.([]interface{})
		if isSlice {
			result[k] = CopyableSlice(slicevalue).DeepCopy()
			continue
		}

		result[k] = v
	}

	return result
}

// DeepCopy will create a deep copy of this slice. The depth of this
// copy is all inclusive. Both maps and slices will be considered when
// making the copy.
func (s CopyableSlice) DeepCopy() []interface{} {
	result := []interface{}{}

	for _, v := range s {
		// Handle maps
		mapvalue, isMap := v.(map[string]interface{})
		if isMap {
			result = append(result, CopyableMap(mapvalue).DeepCopy())
			continue
		}

		// Handle slices
		slicevalue, isSlice := v.([]interface{})
		if isSlice {
			result = append(result, CopyableSlice(slicevalue).DeepCopy())
			continue
		}

		result = append(result, v)
	}

	return result
}

func MergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = MergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

func MergeStringMaps(a, b map[string]string) map[string]string {
	out := make(map[string]string, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

func HaveSameKey(a, b map[string]interface{}) (bool, string) {
	for k := range a {
		_, exists := b[k]
		if exists {
			return true, k
		}
	}
	return false, ""

}
