package isdead

import "math/bits"

const (
	ABCDEFG = 0b11111110_11111110_11111110_11111110_11111110_11111110_11111110_11111110
	BCDEFGH = 0b01111111_01111111_01111111_01111111_01111111_01111111_01111111_01111111
	RANK_8  = 0b11111111_00000000_00000000_00000000_00000000_00000000_00000000_00000000
	RANK_1  = 0b00000000_00000000_00000000_00000000_00000000_00000000_00000000_11111111

	A8 = 0b10000000_00000000_00000000_00000000_00000000_00000000_00000000_00000000
	B8 = 0b01000000_00000000_00000000_00000000_00000000_00000000_00000000_00000000
	G8 = 0b00000010_00000000_00000000_00000000_00000000_00000000_00000000_00000000
	H8 = 0b00000001_00000000_00000000_00000000_00000000_00000000_00000000_00000000
	A7 = 0b00000000_10000000_00000000_00000000_00000000_00000000_00000000_00000000
	H7 = 0b00000000_00000001_00000000_00000000_00000000_00000000_00000000_00000000

	A1 = 0b00000000_00000000_00000000_00000000_00000000_00000000_00000000_10000000
	B1 = 0b00000000_00000000_00000000_00000000_00000000_00000000_00000000_01000000
	G1 = 0b00000000_00000000_00000000_00000000_00000000_00000000_00000000_00000010
	H1 = 0b00000000_00000000_00000000_00000000_00000000_00000000_00000000_00000001
	A2 = 0b00000000_00000000_00000000_00000000_00000000_00000000_10000000_00000000
	H2 = 0b00000000_00000000_00000000_00000000_00000000_00000000_00000001_00000000
)

// TODO also check en passant

type State struct {
	WhiteKings, BlackKings     uint64
	WhitePawns, BlackPawns     uint64
	WhiteBishops, BlackBishops uint64
	Turn                       int // TODO use Turn
}

func computeSingleKingMoves(kings uint64) uint64 {
	return ((kings & BCDEFGH) << 9) |
		(kings << 8) |
		((kings & ABCDEFG) << 7) |
		((kings & BCDEFGH) << 1) |
		((kings & ABCDEFG) >> 1) |
		((kings & BCDEFGH) >> 7) |
		(kings >> 8) |
		((kings & ABCDEFG) >> 9)
}

func computeKingMoves(kings uint64, impossibleSquares uint64) uint64 {
	for {
		newKings := kings | computeSingleKingMoves(kings)
		newKings &^= impossibleSquares
		if newKings == kings {
			return kings
		}
		kings = newKings
	}
}

func computeBishopMoves(bishops uint64, impossibleSquares uint64) uint64 {
	for {
		newBishops := bishops |
			((bishops & BCDEFGH) << 9) |
			((bishops & ABCDEFG) << 7) |
			((bishops & BCDEFGH) >> 7) |
			((bishops & ABCDEFG) >> 9)
		newBishops &^= impossibleSquares
		if newBishops == bishops {
			return bishops
		}
		bishops = newBishops
	}
}

func computeWhitePawnAttack(whitePawns uint64) uint64 {
	return ((whitePawns & BCDEFGH) << 9) | ((whitePawns & ABCDEFG) << 7)
}

func computeBlackPawnAttack(blackPawns uint64) uint64 {
	return ((blackPawns & BCDEFGH) >> 7) | ((blackPawns & ABCDEFG) >> 9)
}

type Result struct {
	Dead       bool
	HasPawns   bool
	HasBishops bool
}

func IsDeadFen(fen string) Result {
	// TODO en passant
	s := State{}
	rank := 7
	file := 0
	part := 0
	hasOtherPieces := false
	for _, c := range fen {
		if c == ' ' {
			part++
			continue
		}
		if part == 0 {
			switch c {
			case '/':
				rank -= 1
				file = 0
			case '1':
				file += 1
			case '2':
				file += 2
			case '3':
				file += 3
			case '4':
				file += 4
			case '5':
				file += 5
			case '6':
				file += 6
			case '7':
				file += 7
			case '8':
				file += 8
			case 'K':
				s.WhiteKings |= 1 << (rank*8 + 7 - file)
				file += 1
			case 'k':
				s.BlackKings |= 1 << (rank*8 + 7 - file)
				file += 1
			case 'B':
				s.WhiteBishops |= 1 << (rank*8 + 7 - file)
				file += 1
			case 'b':
				s.BlackBishops |= 1 << (rank*8 + 7 - file)
				file += 1
			case 'P':
				s.WhitePawns |= 1 << (rank*8 + 7 - file)
				file += 1
			case 'p':
				s.BlackPawns |= 1 << (rank*8 + 7 - file)
				file += 1
			default:
				file += 1
				hasOtherPieces = true
			}
		} else if part == 1 {
			switch c {
			case 'w':
				s.Turn = 0
			case 'b':
				s.Turn = 1
			}
		} else if part == 3 {
			if c >= 'a' && c <= 'h' {
				file = int(c) - 'a'
			}
			if c >= '1' && c <= '8' {
				rank = int(c) - '1'
				if s.Turn == 0 {
					s.BlackPawns |= 1 << (rank*8 + 7 - file)
				} else {
					s.WhitePawns |= 1 << (rank*8 + 7 - file)
				}
			}
		}
	}
	if hasOtherPieces {
		return Result{
			HasPawns:   s.WhitePawns != 0 || s.BlackPawns != 0,
			HasBishops: s.WhiteBishops != 0 || s.BlackBishops != 0,
			Dead:       false,
		}
	}
	return Result{
		HasPawns:   s.WhitePawns != 0 || s.BlackPawns != 0,
		HasBishops: s.WhiteBishops != 0 || s.BlackBishops != 0,
		Dead:       IsDead(s),
	}
}

func IsDead(s State) bool {
	// If there are no pawns or bishops just return
	if s.WhitePawns == 0 && s.BlackPawns == 0 && s.WhiteBishops == 0 && s.BlackBishops == 0 {
		return true
	}
	// PAWNS
	blockedWhitePawns := ((s.WhitePawns << 8) & s.BlackPawns) >> 8
	blockedBlackPawns := ((s.BlackPawns >> 8) & s.WhitePawns) << 8

	whitePawnMoves := s.WhitePawns
	for {
		newWhitePawnMoves := whitePawnMoves | ((whitePawnMoves << 8) &^ blockedBlackPawns)
		if newWhitePawnMoves == whitePawnMoves {
			break
		}
		// FIXME should also check positions where pawns are blocked by own pawns
		whitePawnMoves = newWhitePawnMoves
	}

	blackPawnMoves := s.BlackPawns
	for {
		newBlackPawnMoves := blackPawnMoves | (blackPawnMoves>>8)&^blockedWhitePawns
		if newBlackPawnMoves == blackPawnMoves {
			break
		}
		blackPawnMoves = newBlackPawnMoves
	}

	// Check if any pawns can promote
	if whitePawnMoves&RANK_8 != 0 {
		return false
	}
	if blackPawnMoves&RANK_1 != 0 {
		return false
	}

	whitePawnAttack := computeWhitePawnAttack(whitePawnMoves)
	blackPawnAttack := computeBlackPawnAttack(blackPawnMoves)

	// KINGS
	blockedWhitePawnAttack := computeWhitePawnAttack(blockedWhitePawns)
	blockedBlackPawnAttack := computeBlackPawnAttack(blockedBlackPawns)

	whiteKingMoves := computeKingMoves(s.WhiteKings, blockedBlackPawnAttack|blockedWhitePawns)
	blackKingMoves := computeKingMoves(s.BlackKings, blockedWhitePawnAttack|blockedBlackPawns)

	whiteKingAttack := whiteKingMoves
	blackKingAttack := blackKingMoves

	// BISHOPS
	whiteBishopMoves := computeBishopMoves(s.WhiteBishops, blockedWhitePawns)
	blackBishopMoves := computeBishopMoves(s.BlackBishops, blockedBlackPawns)

	whiteBishopAttack := whiteBishopMoves
	blackBishopAttack := blackBishopMoves

	// Check if the pawns/bishops can capture other pawns/bishops
	if (whitePawnAttack|whiteBishopAttack)&(blackPawnMoves|blackBishopMoves) != 0 {
		return false
	}
	if (blackPawnAttack|blackBishopAttack)&(whitePawnMoves|whiteBishopMoves) != 0 {
		return false
	}

	// Check if the king can capture any pawns
	// (we don't care about bishops as long as they
	// are not blocked)
	if whiteKingAttack&blackPawnMoves != 0 {
		return false
	}
	if blackKingAttack&whitePawnMoves != 0 {
		return false
	}

	// Check if the black king can be mated on a8
	if whiteBishopAttack&A8 != 0 && blackKingMoves&A8 != 0 && blackBishopMoves&B8 != 0 && (blackPawnMoves&A7 != 0 || bits.OnesCount64(s.BlackBishops) > 1) {
		return false
	}
	// Check if the black king can be mated on h8
	if whiteBishopAttack&H8 != 0 && blackKingMoves&H8 != 0 && blackBishopMoves&G8 != 0 && (blackPawnMoves&H7 != 0 || bits.OnesCount64(s.BlackBishops) > 1) {
		return false
	}
	// Check if the white king can be mated on a1
	if blackBishopAttack&A1 != 0 && whiteKingMoves&A1 != 0 && whiteBishopMoves&B1 != 0 && (whitePawnMoves&A2 != 0 || bits.OnesCount64(s.WhiteBishops) > 1) {
		return false
	}
	// Check if the white king can be mated on h1
	if blackBishopAttack&H1 != 0 && whiteKingMoves&H1 != 0 && whiteBishopMoves&G1 != 0 && (whitePawnMoves&H2 != 0 || bits.OnesCount64(s.WhiteBishops) > 1) {
		return false
	}

	return true
}
