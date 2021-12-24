package ipmi

import "fmt"

type GetSDRRequest struct {
	ReservationID uint16 // LS Byte first
	RecordID      uint16 // LS Byte first
	OffsetInfo    uint8
	BytesToRead   uint8 // FFh means read entire record
}

type GetSDRResponse struct {
	NextRecordID uint16
	RecordData   []byte
}

func (req *GetSDRRequest) Pack() []byte {
	msg := make([]byte, 6)
	packUint16L(req.ReservationID, msg, 0)
	packUint16L(req.RecordID, msg, 2)
	packUint8(req.OffsetInfo, msg, 4)
	packUint8(req.BytesToRead, msg, 5)
	return msg
}

func (req *GetSDRRequest) Command() Command {
	return CommandGetSDR
}

func (res *GetSDRResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShort
	}
	res.NextRecordID, _, _ = unpackUint16L(msg, 0)
	res.RecordData, _, _ = unpackBytes(msg, 2, len(msg)-2)
	return nil
}

func (res *GetSDRResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetSDRResponse) Format() string {
	return fmt.Sprintf("%v", res)
}

func (c *Client) GetSDR(recordID uint16) (response *GetSDRResponse, err error) {
	request := &GetSDRRequest{
		ReservationID: 0,
		RecordID:      recordID,
		OffsetInfo:    0,
		BytesToRead:   0xff,
	}
	response = &GetSDRResponse{}
	err = c.Exchange(request, response)
	return
}

// GetSDRs fetches the SDR records of the specified RecordType.
// Pass 0 to fetch all SDR records.
func (c *Client) GetSDRs(recordType SDRRecordType) ([]*SDR, error) {
	var recordID uint16 = 0

	var out = make([]*SDR, 0)
	for {
		res, err := c.GetSDR(recordID)
		if err != nil {
			return nil, fmt.Errorf("GetSDR for recordID (%#0x) failed, err: %s", recordID, err)
		}
		recordID = res.NextRecordID

		sdr, err := ParseSDR(res.RecordData, res.NextRecordID)
		if err != nil {
			return nil, fmt.Errorf("ParseSDR for recordID (%#0x) failed, err: %s", recordID, err)
		}
		if recordType != 0 && sdr.RecordHeader.RecordType != recordType {
			continue
		}
		out = append(out, sdr)

		if recordID == 0xffff {
			break
		}
	}

	return out, nil
}
