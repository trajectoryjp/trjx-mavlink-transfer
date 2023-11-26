package buffer

// serverと共有コード
// trjxTransferと共通コード

import (
	"errors"
	"log"
)

type Buffer struct {
	buffer  []byte
	indexFE indexDef
	indexFD indexDef
}

type indexDef struct {
	sync bool
}

func CreateMavBuffer() *Buffer {
	bufferObj := Buffer{
		buffer: make([]byte, 0),
	}
	return &bufferObj
}

func (bf *Buffer) Push(b []byte) {
	bf.buffer = append(bf.buffer, b...)
}

func (bf *Buffer) Pop() (packet []byte) {
	if bf.indexFE.sync {
		return bf.searchFE()

	} else if bf.indexFD.sync {
		return bf.searchFD()

	} else {
		if packet = bf.searchFE(); packet == nil {
			if packet = bf.searchFD(); packet == nil {
				bf.dropBuffer()
			}
		}
		return packet
	}
}

func (bf *Buffer) searchFE() (packet []byte) {
	blen := len(bf.buffer)
	if blen >= 8 {
		for k, b := range bf.buffer[0 : blen-8] {
			if b == 0xFE {
				// packet長
				if packetLen, complete, err := packetLen(bf.buffer[k:]); err == nil {
					if complete {
						nextIndex := k + packetLen
						if nextIndex < blen-1 {
							if bf.buffer[nextIndex] == 0xFE {
								if !bf.indexFE.sync {
									log.Print("v1 sync - continue")
								}

								bf.indexFE.sync = true
								bf.indexFD.sync = false
								packet = bf.buffer[k:nextIndex]
								bf.buffer = bf.buffer[nextIndex:]
								return packet
							}
							// 継続

						} else if nextIndex == blen {
							if !bf.indexFE.sync {
								log.Print("v1 sync")
							}

							bf.indexFE.sync = true
							bf.indexFD.sync = false
							packet = bf.buffer[k:nextIndex]
							bf.buffer = []byte{}
							return packet
						}

					} else if packetLen >= 8 {
						return nil
					}

				} else {
					log.Fatal("unexpceted error searchFE")
				}
			}
		}
	}
	return nil
}

func (bf *Buffer) searchFD() (packet []byte) {
	blen := len(bf.buffer)
	if blen >= 12 {
		for k, b := range bf.buffer[0 : blen-12] {
			if b == 0xFD {
				// packet長
				if packetLen, complete, err := packetLen(bf.buffer[k:]); err == nil {
					if complete {
						nextIndex := k + packetLen
						if nextIndex < blen-1 {
							if bf.buffer[nextIndex] == 0xFD {
								if !bf.indexFD.sync {
									log.Print("v2 sync - continue")
								}

								bf.indexFE.sync = false
								bf.indexFD.sync = true
								packet = bf.buffer[k:nextIndex]
								bf.buffer = bf.buffer[nextIndex:]
								return packet
							}
							// 継続

						} else if nextIndex == blen {
							if !bf.indexFD.sync {
								log.Print("v2 sync")
							}

							bf.indexFE.sync = false
							bf.indexFD.sync = true
							packet = bf.buffer[k:nextIndex]
							bf.buffer = []byte{}
							return packet
						}

					} else if packetLen >= 12 {
						return nil
					}

				} else {
					log.Fatal("unexpceted error searchFD")
				}
			}
		}
	}
	return nil
}

func (bf *Buffer) dropBuffer() {
	if len(bf.buffer) > 600 {
		log.Print("dropBuffer")
		bf.buffer = bf.buffer[300:]
		bf.indexFE.sync = false
		bf.indexFD.sync = false
	}
}

// bufの0番目はSTXであること
// boolは
//  - true：buf内にパケットを含む
//  - false：buf内にすべてのパケットは含まれない
func packetLen(buf []byte) (paketLen int, complete bool, err error) {
	blen := len(buf)

	switch buf[0] {
	case 0xFE:
		if blen >= 2 {
			paketLen = int(buf[1]) + 8
			if paketLen <= 263 {
				if blen >= paketLen {
					return paketLen, true, nil

				} else {
					return paketLen, false, nil
				}

			} else {
				// 規定長以上
				return -2, false, nil
			}

		} else {
			return -1, false, nil
		}

	case 0xFD:
		if blen >= 3 {
			paketLen = int(buf[1]) + 12
			if buf[2] == 0x01 { // INC-FLAGS
				paketLen += 13
			}
			if paketLen <= 280 {
				if blen >= paketLen {
					return paketLen, true, nil

				} else {
					return paketLen, false, nil
				}

			} else {
				// 規定長以上
				return -2, false, nil
			}

		} else {
			return -1, false, nil
		}

	default:
		return -1, false, errors.New("STX error")
	}
}
