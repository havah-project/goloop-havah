/*
 * Copyright 2020 ICON Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package state

import (
	"reflect"
	"testing"
)

func Test_getNextID(t *testing.T) {
	type args struct {
		id []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "ZeroLeading1",
			args: args{[]byte{0x00, 0x01}},
			want: []byte{0x00, 0x02},
		},
		{
			name: "ZeroLeading2",
			args: args{[]byte{0x00, 0xff}},
			want: []byte{0x01, 0x00},
		},
		{
			name: "Overflow",
			args: args{[]byte{0xff, 0xff}},
			want: []byte{0x00, 0x00},
		},
		{
			name: "Normal",
			args: args{[]byte{0xab, 0xff}},
			want: []byte{0xac, 0x00},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getNextID(tt.args.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getNextID() = %v, want %v", got, tt.want)
			}
		})
	}
}