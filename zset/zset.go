//Package zset is a port of t_zset.c in Redis
/*
 * Copyright (c) 2009-2012, Salvatore Sanfilippo <antirez at gmail dot com>
 * Copyright (c) 2009-2012, Pieter Noordhuis <pcnoordhuis at gmail dot com>
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 *   * Redistributions of source code must retain the above copyright notice,
 *     this list of conditions and the following disclaimer.
 *   * Redistributions in binary form must reproduce the above copyright
 *     notice, this list of conditions and the following disclaimer in the
 *     documentation and/or other materials provided with the distribution.
 *   * Neither the name of Redis nor the names of its contributors may be used
 *     to endorse or promote products derived from this software without
 *     specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE
 * LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 */
package zset

import (
	"math/rand"
	"sync"
)

type Number interface {
	~int8 | ~uint8 | ~int | ~uint | ~int16 | ~uint16 | ~int32 | ~uint32 | ~int64 | ~uint64 | ~float32 | ~float64
}

type Comparable interface {
	Number | ~string
}

const zSkiplistMaxlevel = 32

type (
	skipListLevel[C Comparable, N Number] struct {
		forward *skipListNode[C, N]
		span    uint64
	}

	skipListNode[C Comparable, N Number] struct {
		objID    C
		score    N
		backward *skipListNode[C, N]
		level    []*skipListLevel[C, N]
	}
	obj[T any, C Comparable, N Number] struct {
		key        C
		attachment T // 使用范型
		score      N
	}

	skipList[T any, C Comparable, N Number] struct {
		header *skipListNode[C, N]
		tail   *skipListNode[C, N]
		length int64
		level  int16
	}
	// SortedSet is the final exported sorted set we can use
	SortedSet[T any, C Comparable, N Number] struct {
		dict map[C]*obj[T, C, N]
		zsl  *skipList[T, C, N]
		sync.RWMutex
	}
	zrangespec[N Number] struct {
		min   N
		max   N
		minex int32
		maxex int32
	}
	zlexrangespec[C Comparable] struct {
		minKey C
		maxKey C
		minex  int
		maxex  int
	}
)

func zslCreateNode[C Comparable, N Number](level int16, score N, id C) *skipListNode[C, N] {
	n := &skipListNode[C, N]{
		score: score,
		objID: id,
		level: make([]*skipListLevel[C, N], level),
	}
	for i := range n.level {
		n.level[i] = new(skipListLevel[C, N])
	}
	return n
}

func zslCreate[T any, C Comparable, N Number]() *skipList[T, C, N] {
	var id C
	var score N
	return &skipList[T, C, N]{
		level:  1,
		header: zslCreateNode(zSkiplistMaxlevel, score, id),
	}
}

const zSkiplistP = 0.25 /* Skiplist P = 1/4 */

/* Returns a random level for the new skiplist node we are going to create.
 * The return value of this function is between 1 and _ZSKIPLIST_MAXLEVEL
 * (both inclusive), with a powerlaw-alike distribution where higher
 * levels are less likely to be returned. */
func randomLevel() int16 {
	level := int16(1)
	for float32(rand.Int31()&0xFFFF) < (zSkiplistP * 0xFFFF) {
		level++
	}
	if level < zSkiplistMaxlevel {
		return level
	}
	return zSkiplistMaxlevel
}

/* zslInsert a new node in the skiplist. Assumes the element does not already
 * exist (up to the caller to enforce that). The skiplist takes ownership
 * of the passed SDS string 'obj'. */
func (zsl *skipList[T, C, N]) zslInsert(score N, id C) *skipListNode[C, N] {
	update := make([]*skipListNode[C, N], zSkiplistMaxlevel)
	rank := make([]uint64, zSkiplistMaxlevel)
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		/* store rank that is crossed to reach the insert position */
		if i == zsl.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}
		if x.level[i] != nil {
			for x.level[i].forward != nil &&
				(x.level[i].forward.score < score ||
					(x.level[i].forward.score == score && x.level[i].forward.objID < id)) {
				rank[i] += x.level[i].span
				x = x.level[i].forward
			}
		}
		update[i] = x
	}
	/* we assume the element is not already inside, since we allow duplicated
	 * scores, reinserting the same element should never happen since the
	 * caller of zslInsert() should test in the hash table if the element is
	 * already inside or not. */
	level := randomLevel()
	if level > zsl.level {
		for i := zsl.level; i < level; i++ {
			rank[i] = 0
			update[i] = zsl.header
			update[i].level[i].span = uint64(zsl.length)
		}
		zsl.level = level
	}
	x = zslCreateNode(level, score, id)
	for i := int16(0); i < level; i++ {
		x.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = x

		/* update span covered by update[i] as x is inserted here */
		x.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = (rank[0] - rank[i]) + 1
	}

	/* increment span for untouched levels */
	for i := level; i < zsl.level; i++ {
		update[i].level[i].span++
	}

	if update[0] == zsl.header {
		x.backward = nil
	} else {
		x.backward = update[0]

	}
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x
	} else {
		zsl.tail = x
	}
	zsl.length++
	return x
}

/* Internal function used by zslDelete, zslDeleteByScore and zslDeleteByRank */
func (zsl *skipList[T, C, N]) zslDeleteNode(x *skipListNode[C, N], update []*skipListNode[C, N]) {
	for i := int16(0); i < zsl.level; i++ {
		if update[i].level[i].forward == x {
			update[i].level[i].span += x.level[i].span - 1
			update[i].level[i].forward = x.level[i].forward
		} else {
			update[i].level[i].span--
		}
	}
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x.backward
	} else {
		zsl.tail = x.backward
	}
	for zsl.level > 1 && zsl.header.level[zsl.level-1].forward == nil {
		zsl.level--
	}
	zsl.length--
}

/* Delete an element with matching score/element from the skiplist.
 * The function returns 1 if the node was found and deleted, otherwise
 * 0 is returned.
 *
 * If 'node' is NULL the deleted node is freed by zslFreeNode(), otherwise
 * it is not freed (but just unlinked) and *node is set to the node pointer,
 * so that it is possible for the caller to reuse the node (including the
 * referenced SDS string at node->obj). */
func (zsl *skipList[T, C, N]) zslDelete(score N, id C) int {
	update := make([]*skipListNode[C, N], zSkiplistMaxlevel)
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			(x.level[i].forward.score < score ||
				(x.level[i].forward.score == score &&
					x.level[i].forward.objID < id)) {
			x = x.level[i].forward
		}
		update[i] = x
	}
	/* We may have multiple elements with the same score, what we need
	 * is to find the element with both the right score and object. */
	x = x.level[0].forward
	if x != nil && score == x.score && x.objID == id {
		zsl.zslDeleteNode(x, update)
		return 1
	}
	return 0 /* not found */
}

func zslValueGteMin[N Number](value N, spec *zrangespec[N]) bool {
	if spec.minex != 0 {
		return value > spec.min
	}
	return value >= spec.min
}

func zslValueLteMax[N Number](value N, spec *zrangespec[N]) bool {
	if spec.maxex != 0 {
		return value < spec.max
	}
	return value <= spec.max
}

/* Returns if there is a part of the zset is in range. */
func (zsl *skipList[T, C, N]) zslIsInRange(ran *zrangespec[N]) bool {
	/* Test for ranges that will always be empty. */
	if ran.min > ran.max ||
		(ran.min == ran.max && (ran.minex != 0 || ran.maxex != 0)) {
		return false
	}
	x := zsl.tail
	if x == nil || !zslValueGteMin(x.score, ran) {
		return false
	}
	x = zsl.header.level[0].forward
	if x == nil || !zslValueLteMax(x.score, ran) {
		return false
	}
	return true
}

/* Find the first node that is contained in the specified range.
 * Returns NULL when no element is contained in the range. */
func (zsl *skipList[T, C, N]) zslFirstInRange(ran *zrangespec[N]) *skipListNode[C, N] {
	/* If everything is out of range, return early. */
	if !zsl.zslIsInRange(ran) {
		return nil
	}

	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		/* Go forward while *OUT* of range. */
		for x.level[i].forward != nil &&
			!zslValueGteMin(x.level[i].forward.score, ran) {
			x = x.level[i].forward
		}
	}
	/* This is an inner range, so the next node cannot be NULL. */
	x = x.level[0].forward
	//serverAssert(x != NULL);

	/* Check if score <= max. */
	if !zslValueLteMax(x.score, ran) {
		return nil
	}
	return x
}

/* Find the last node that is contained in the specified range.
 * Returns NULL when no element is contained in the range. */
func (zsl *skipList[T, C, N]) zslLastInRange(ran *zrangespec[N]) *skipListNode[C, N] {

	/* If everything is out of range, return early. */
	if !zsl.zslIsInRange(ran) {
		return nil
	}
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		/* Go forward while *IN* range. */
		for x.level[i].forward != nil &&
			zslValueLteMax(x.level[i].forward.score, ran) {
			x = x.level[i].forward
		}
	}
	/* This is an inner range, so this node cannot be NULL. */
	//serverAssert(x != NULL);

	/* Check if score >= min. */
	if !zslValueGteMin(x.score, ran) {
		return nil
	}
	return x
}

/* Delete all the elements with score between min and max from the skiplist.
 * Min and max are inclusive, so a score >= min || score <= max is deleted.
 * Note that this function takes the reference to the hash table view of the
 * sorted set, in order to remove the elements from the hash table too. */
func (zsl *skipList[T, C, N]) zslDeleteRangeByScore(ran *zrangespec[N], dict map[C]*obj[T, C, N]) uint64 {
	removed := uint64(0)
	update := make([]*skipListNode[C, N], zSkiplistMaxlevel)
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil {
			var condition bool
			if ran.minex != 0 {
				condition = x.level[i].forward.score <= ran.min
			} else {
				condition = x.level[i].forward.score < ran.min
			}
			if !condition {
				break
			}
			x = x.level[i].forward
		}
		update[i] = x
	}

	/* Current node is the last with score < or <= min. */
	x = x.level[0].forward

	/* Delete nodes while in range. */
	for x != nil {
		var condition bool
		if ran.maxex != 0 {
			condition = x.score < ran.max
		} else {
			condition = x.score <= ran.max
		}
		if !condition {
			break
		}
		next := x.level[0].forward
		zsl.zslDeleteNode(x, update)
		delete(dict, x.objID)
		// Here is where x->obj is actually released.
		// And golang has GC, don't need to free manually anymore
		//zslFreeNode(x)
		removed++
		x = next
	}
	return removed
}

func (zsl *skipList[T, C, N]) zslDeleteRangeByLex(ran *zlexrangespec[C], dict map[C]*obj[T, C, N]) uint64 {
	removed := uint64(0)

	update := make([]*skipListNode[C, N], zSkiplistMaxlevel)
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && !zslLexValueGteMin(x.level[i].forward.objID, ran) {
			x = x.level[i].forward
		}
		update[i] = x
	}

	/* Current node is the last with score < or <= min. */
	x = x.level[0].forward

	/* Delete nodes while in range. */
	for x != nil && zslLexValueLteMax(x.objID, ran) {
		next := x.level[0].forward
		zsl.zslDeleteNode(x, update)
		delete(dict, x.objID)
		removed++
		x = next
	}
	return removed
}

func zslLexValueGteMin[C Comparable](id C, spec *zlexrangespec[C]) bool {
	if spec.minex != 0 {
		return compareKey(id, spec.minKey) > 0
	}
	return compareKey(id, spec.minKey) >= 0
}

func compareKey[C Comparable](a, b C) int8 {
	if a == b {
		return 0
	} else if a > b {
		return 1
	}
	return -1
}

func zslLexValueLteMax[C Comparable](id C, spec *zlexrangespec[C]) bool {
	if spec.maxex != 0 {
		return compareKey(id, spec.maxKey) < 0
	}
	return compareKey(id, spec.maxKey) <= 0
}

/* Delete all the elements with rank between start and end from the skiplist.
 * Start and end are inclusive. Note that start and end need to be 1-based */
func (zsl *skipList[T, C, N]) zslDeleteRangeByRank(start, end uint64, dict map[C]*obj[T, C, N]) uint64 {
	update := make([]*skipListNode[C, N], zSkiplistMaxlevel)
	var traversed, removed uint64

	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && (traversed+x.level[i].span) < start {
			traversed += x.level[i].span
			x = x.level[i].forward
		}
		update[i] = x
	}

	traversed++
	x = x.level[0].forward
	for x != nil && traversed <= end {
		next := x.level[0].forward
		zsl.zslDeleteNode(x, update)
		delete(dict, x.objID)
		removed++
		traversed++
		x = next
	}
	return removed
}

/* Find the rank for an element by both score and obj.
 * Returns 0 when the element cannot be found, rank otherwise.
 * Note that the rank is 1-based due to the span of zsl->header to the
 * first element. */
func (zsl *skipList[T, C, N]) zslGetRank(score N, key C) int64 {
	rank := uint64(0)
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			(x.level[i].forward.score < score ||
				(x.level[i].forward.score == score &&
					x.level[i].forward.objID <= key)) {
			rank += x.level[i].span
			x = x.level[i].forward
		}

		/* x might be equal to zsl->header, so test if obj is non-NULL */
		if x.objID == key {
			return int64(rank)
		}
	}
	return 0
}

/* Finds an element by its rank. The rank argument needs to be 1-based. */
func (zsl *skipList[T, C, N]) zslGetElementByRank(rank uint64) *skipListNode[C, N] {
	traversed := uint64(0)
	x := zsl.header
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && (traversed+x.level[i].span) <= rank {
			traversed += x.level[i].span
			x = x.level[i].forward
		}
		if traversed == rank {
			return x
		}
	}
	return nil
}

/*-----------------------------------------------------------------------------
 * Common sorted set API
 *----------------------------------------------------------------------------*/

// New creates a new SortedSet and return its pointer
func New[T any, C Comparable, N Number]() *SortedSet[T, C, N] {
	s := &SortedSet[T, C, N]{
		dict: make(map[C]*obj[T, C, N]),
		zsl:  zslCreate[T, C, N](),
	}
	return s
}

// Length returns counts of elements
func (z *SortedSet[T, C, N]) Length() int64 {
	z.RLock()
	defer z.RUnlock()
	return z.zsl.length
}

// Set is used to add or update an element
func (z *SortedSet[T, C, N]) Set(key C, score N, dat T) {
	z.Lock()
	defer z.Unlock()

	v, ok := z.dict[key]
	z.dict[key] = &obj[T, C, N]{attachment: dat, key: key, score: score}
	if ok {
		/* Remove and re-insert when score changes. */
		if score != v.score {
			z.zsl.zslDelete(v.score, key)
			z.zsl.zslInsert(score, key)
		}
	} else {
		z.zsl.zslInsert(score, key)
	}
}

// IncrBy ..
func (z *SortedSet[T, C, N]) Incr(key C, score N) (N, T) {
	z.Lock()
	defer z.Unlock()
	v, ok := z.dict[key]
	if !ok {

		var t T
		// use negative infinity ?
		return 0, t
	}
	if score != 0 {
		z.zsl.zslDelete(v.score, key)
		v.score += score
		z.zsl.zslInsert(v.score, key)
	}
	return v.score, v.attachment
}

// Delete removes an element from the SortedSet
// by its key.
func (z *SortedSet[T, C, N]) Delete(key C) (ok bool) {
	z.Lock()
	defer z.Unlock()

	v, ok := z.dict[key]
	if ok {
		z.zsl.zslDelete(v.score, key)
		delete(z.dict, key)
		return true
	}
	return false
}

// GetRank returns position,score and extra data of an element which
// found by the parameter key.
// The parameter reverse determines the rank is descent or ascend，
// true means descend and false means ascend.
func (z *SortedSet[T, C, N]) GetRank(key C, reverse bool) (rank int64, score N, data T) {
	z.RLock()
	defer z.RUnlock()

	v, ok := z.dict[key]
	if !ok {
		var t T
		return -1, 0, t
	}
	r := z.zsl.zslGetRank(v.score, key)
	if reverse {
		r = z.zsl.length - r
	} else {
		r--
	}
	return int64(r), v.score, v.attachment

}

// GetData returns data stored in the map by its key
func (z *SortedSet[T, C, N]) GetData(key C) (data T, ok bool) {
	z.RLock()
	defer z.RUnlock()

	o, ok := z.dict[key]
	if !ok {
		var t T
		return t, false
	}
	return o.attachment, true
}

// GetScore implements ZScore
func (z *SortedSet[T, C, N]) GetScore(key C) (score N, ok bool) {
	z.RLock()
	defer z.RUnlock()

	o, ok := z.dict[key]
	if !ok {
		return 0, false
	}
	return o.score, true
}

// GetDataByRank returns the id,score and extra data of an element which
// found by position in the rank.
// The parameter rank is the position, reverse says if in the descend rank.
func (z *SortedSet[T, C, N]) GetDataByRank(rank int64, reverse bool) (key C, score N, data T) {
	z.RLock()
	defer z.RUnlock()

	if rank < 0 || rank > z.zsl.length {
		return key, score, data
	}
	if reverse {
		rank = z.zsl.length - rank
	} else {
		rank++
	}
	n := z.zsl.zslGetElementByRank(uint64(rank))
	if n == nil {
		return key, score, data
	}
	dat := z.dict[n.objID]
	if dat == nil {
		return key, score, data
	}
	return dat.key, dat.score, dat.attachment
}

// Range implements ZRANGE
// RevRange implements ZREVRANGE
func (z *SortedSet[T, C, N]) Range(start, end int64, reverse bool, f func(N, C, T)) {
	z.RLock()
	defer z.RUnlock()
	z.commonRange(start, end, reverse, f)
}

// Range by score
func (z *SortedSet[T, C, N]) RangeByScore(min, max N, reverse bool, f func(N, C, T)) {
	z.RLock()
	defer z.RUnlock()
	z.scoreRange(min, max, reverse, f)
}

// Range by score
func (z *SortedSet[T, C, N]) scoreRange(min, max N, reverse bool, f func(N, C, T)) {
	var node *skipListNode[C, N]
	var zran = &zrangespec[N]{min: min, max: max}
	if reverse {
		node = z.zsl.zslLastInRange(zran)
	} else {
		node = z.zsl.zslFirstInRange(zran)
	}
	if node == nil {
		return
	}

	for node != nil {
		if (reverse && node.score < min) || node.score > max {
			return
		}
		f(node.score, node.objID, z.dict[node.objID].attachment)
		if reverse {
			node = node.backward
		} else {
			node = node.level[0].forward
		}
	}
}

func (z *SortedSet[T, C, N]) commonRange(start, end int64, reverse bool, f func(N, C, T)) {
	l := z.zsl.length
	if start < 0 {
		start += l
		if start < 0 {
			start = 0
		}
	}
	if end < 0 {
		end += l
	}

	if start > end || start >= l {
		return
	}
	if end >= l {
		end = l - 1
	}
	span := (end - start) + 1

	var node *skipListNode[C, N]
	if reverse {
		node = z.zsl.tail
		if start > 0 {
			node = z.zsl.zslGetElementByRank(uint64(l - start))
		}
	} else {
		node = z.zsl.header.level[0].forward
		if start > 0 {
			node = z.zsl.zslGetElementByRank(uint64(start + 1))
		}
	}
	for span > 0 {
		span--
		k := node.objID
		s := node.score
		f(s, k, z.dict[k].attachment)
		if reverse {
			node = node.backward
		} else {
			node = node.level[0].forward
		}
	}
}
