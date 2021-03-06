// Copyright 2017-2021 DERO Project. All rights reserved.
// Use of this source code in any form is governed by RESEARCH license.
// license can be found in the LICENSE file.
// GPG: 0F39 E425 8C65 3947 702A  8234 08B2 0360 A03A 9DE8
//
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY
// EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL
// THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
// PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT,
// STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF
// THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package main

import "fmt"
import "context"

//import "encoding/hex"
import "runtime/debug"

//import "github.com/romana/rlog"
import "github.com/deroproject/derohe/cryptography/crypto"
import "github.com/deroproject/derohe/config"
import "github.com/deroproject/derohe/rpc"
import "github.com/deroproject/derohe/dvm"

//import "github.com/deroproject/derohe/transaction"
import "github.com/deroproject/derohe/blockchain"

import "github.com/deroproject/graviton"

func (DERO_RPC_APIS) GetSC(ctx context.Context, p rpc.GetSC_Params) (result rpc.GetSC_Result, err error) {

	defer func() { // safety so if anything wrong happens, we return error
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occured. stack trace %s", debug.Stack())
		}
	}()

	scid := crypto.HashHexToHash(p.SCID)

	topoheight := chain.Load_TOPO_HEIGHT()

	if p.TopoHeight >= 1 {
		topoheight = p.TopoHeight
	}

	toporecord, err := chain.Store.Topo_store.Read(topoheight)
	// we must now fill in compressed ring members
	if err == nil {
		var ss *graviton.Snapshot
		ss, err = chain.Store.Balance_store.LoadSnapshot(toporecord.State_Version)
		if err == nil {
			var sc_meta_tree *graviton.Tree
			if sc_meta_tree, err = ss.GetTree(config.SC_META); err == nil {
				var meta_bytes []byte
				if meta_bytes, err = sc_meta_tree.Get(blockchain.SC_Meta_Key(scid)); err == nil {
					var meta blockchain.SC_META_DATA
					if err = meta.UnmarshalBinary(meta_bytes); err == nil {
						result.Balance = meta.Balance
					}
				}
			} else {
				return
			}

			if sc_data_tree, err := ss.GetTree(string(scid[:])); err == nil {
				if p.Code { // give SC code
					var code_bytes []byte
					var v dvm.Variable
					if code_bytes, err = sc_data_tree.Get(blockchain.SC_Code_Key(scid)); err == nil {
						if err = v.UnmarshalBinary(code_bytes); err != nil {
							result.Code = "Unmarshal error"
						} else {
							result.Code = v.Value.(string)
						}
					}
				}

				// give any uint64 keys data if any
				for _, value := range p.KeysUint64 {
					var v dvm.Variable
					key, _ := dvm.Variable{Type: dvm.Uint64, Value: value}.MarshalBinary()

					var value_bytes []byte
					if value_bytes, err = sc_data_tree.Get(key); err != nil {
						result.ValuesUint64 = append(result.ValuesUint64, fmt.Sprintf("NOT AVAILABLE err: %s", err))
						continue
					}
					if err = v.UnmarshalBinary(value_bytes); err != nil {
						result.ValuesUint64 = append(result.ValuesUint64, "Unmarshal error")
						continue
					}
					switch v.Type {
					case dvm.Uint64:
						result.ValuesUint64 = append(result.ValuesUint64, fmt.Sprintf("%d", v.Value))
					case dvm.String:
						result.ValuesUint64 = append(result.ValuesUint64, fmt.Sprintf("%s", v.Value))
					default:
						result.ValuesUint64 = append(result.ValuesUint64, "UNKNOWN Data type")
					}
				}
				for _, value := range p.KeysString {
					var v dvm.Variable
					key, _ := dvm.Variable{Type: dvm.String, Value: value}.MarshalBinary()

					var value_bytes []byte
					if value_bytes, err = sc_data_tree.Get(key); err != nil {
						fmt.Printf("Getting key %x\n", key)
						result.ValuesString = append(result.ValuesString, fmt.Sprintf("NOT AVAILABLE err: %s", err))
						continue
					}
					if err = v.UnmarshalBinary(value_bytes); err != nil {
						result.ValuesString = append(result.ValuesString, "Unmarshal error")
						continue
					}
					switch v.Type {
					case dvm.Uint64:
						result.ValuesString = append(result.ValuesUint64, fmt.Sprintf("%d", v.Value))
					case dvm.String:
						result.ValuesString = append(result.ValuesString, fmt.Sprintf("%s", v.Value))
					default:
						result.ValuesString = append(result.ValuesString, "UNKNOWN Data type")
					}
				}

				for _, value := range p.KeysBytes {
					var v dvm.Variable
					key, _ := dvm.Variable{Type: dvm.String, Value: string(value)}.MarshalBinary()

					var value_bytes []byte
					if value_bytes, err = sc_data_tree.Get(key); err != nil {
						result.ValuesBytes = append(result.ValuesBytes, "NOT AVAILABLE")
						continue
					}
					if err = v.UnmarshalBinary(value_bytes); err != nil {
						result.ValuesBytes = append(result.ValuesBytes, "Unmarshal error")
						continue
					}
					switch v.Type {
					case dvm.Uint64:
						result.ValuesBytes = append(result.ValuesBytes, fmt.Sprintf("%d", v.Value))
					case dvm.String:
						result.ValuesBytes = append(result.ValuesBytes, fmt.Sprintf("%s", v.Value))
					default:
						result.ValuesBytes = append(result.ValuesBytes, "UNKNOWN Data type")
					}
				}

			}

		}

	}

	result.Status = "OK"
	err = nil

	//logger.Debugf("result %+v\n", result);
	return
}
