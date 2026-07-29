package bankprovider

import (
	"context"
	"fmt"

	"sumni-finance-backend/internal/treasury/domain"
)

var mockBanks = func() []domain.BankInfo {
	raw := []struct {
		name, code, shortName, logo string
		bin                         string
	}{
		{"Ngân hàng TMCP Ngoại Thương Việt Nam", "VCB", "Vietcombank", "https://api.vietqr.io/img/VCB.png", "970436"},
		{"Ngân hàng TMCP Á Châu", "ACB", "ACB", "https://api.vietqr.io/img/ACB.png", "970416"},
		{"Ngân hàng TMCP Kỹ Thương Việt Nam", "TCB", "Techcombank", "https://api.vietqr.io/img/TCB.png", "970407"},
		{"Ngân hàng TMCP Tiên Phong", "TPB", "TPBank", "https://api.vietqr.io/img/TPB.png", "970423"},
		{"Ngân hàng TMCP An Bình", "ABB", "ABBANK", "https://api.vietqr.io/img/ABB.png", "970425"},
	}

	banks := make([]domain.BankInfo, 0, len(raw))
	for _, r := range raw {
		b, err := domain.NewBankInfo(r.name, r.bin, r.code, r.shortName, r.logo)
		if err != nil {
			panic(fmt.Sprintf("lookuptestutils: invalid mock bank %q: %v", r.code, err))
		}
		banks = append(banks, b)
	}
	return banks
}()

type StubClient struct {
	banks []domain.BankInfo
	err   error
}

func NewStubClient() *StubClient {
	return &StubClient{banks: mockBanks}
}

func NewStubClientWithError(err error) *StubClient {
	return &StubClient{err: err}
}

func (s *StubClient) FindBankInfoByCode(_ context.Context, bankCode string) (domain.BankInfo, error) {
	if s.err != nil {
		return domain.BankInfo{}, s.err
	}

	for _, b := range s.banks {
		if b.BankCode() == bankCode {
			return b, nil
		}
	}

	return domain.BankInfo{}, fmt.Errorf("bank with code %q not found", bankCode)
}
