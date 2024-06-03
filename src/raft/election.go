package raft

import (
	"math/rand"
	"sync"
	"time"
)

func (rf *Raft) resetElectionTimer() {
	t := time.Now()
	electionTimeout := time.Duration(150+rand.Intn(150)) * time.Millisecond
	rf.electionTime = t.Add(electionTimeout)
}

// for follower to restart a new leader election
func (rf *Raft) electLeader() {
	// when step in this function, rf.mu has locked in ticker

	// firstly append currentTerm
	rf.currentTerm++

	// transformed state to candidate
	rf.state = StateCandidate

	//candidate rule: always vote for self
	rf.votedFor = rf.me

	rf.persist()
	rf.resetElectionTimer()

	term := rf.currentTerm
	voteCounter := 1

	lastLog := rf.log.lastLog()
	DPrintf("[%v]: Start a new election term %d\n", rf.me, term)

	args := RequestVoteArgs{
		CandidateId:  rf.me,
		Term:         term,
		LastLogIndex: lastLog.Index,
		LastLogTerm:  lastLog.Term,
	}

	var becomeLeader sync.Once

	for peer, _ := range rf.peers {
		if peer == rf.me {
			continue
		}

		go rf.candidateRequestVote(peer, &args, &voteCounter, &becomeLeader)
	}
}

func (rf *Raft) setNewTerm(term int) {
	if term > rf.currentTerm || rf.currentTerm == 0 {
		rf.state = StateFollower
		rf.currentTerm = term
		rf.votedFor = -1
		DPrintf("[%d]: set term %v\n", rf.me, rf.currentTerm)
		rf.persist()
	}
}

func (rf *Raft) candidateRequestVote(serverId int, args *RequestVoteArgs, voteCounter *int, becomeLeader *sync.Once) {
	DPrintf("[%v]: term %v send request vote to %d\n", rf.me, rf.currentTerm, serverId)

	reply := RequestVoteReply{}
	ok := rf.sendRequestVote(serverId, args, &reply)
	if !ok {
		return
	}

	rf.mu.Lock()
	defer rf.mu.Unlock()

	// other peer's term newer then current node
	if reply.Term > args.Term {
		DPrintf("[%v]: term 更新, 自己变成follower并更新term\n", rf.me)
		rf.setNewTerm(reply.Term)
		return
	}

	// when other node receive a request vote, if other node's term < args.term
	// they will firstly setNewTerm(args.Term)
	// so if reply.Term < args.Term, the reply is from a pre term vote request
	if reply.Term < args.Term {
		DPrintf("[%d]: %d 的term %d 已经失效，结束\n", rf.me, serverId, reply.Term)
		return
	}

	if !reply.Grant {
		DPrintf("[%v]: %d 拒绝投票给自己\n", rf.me, serverId)
		return
	}

	DPrintf("[%v]: %d 投票给自己\n", rf.me, serverId)

	*voteCounter++

	// receiver majority servers' vote and double check self state
	if *voteCounter > len(rf.peers)/2 && rf.currentTerm == args.Term && rf.state == StateCandidate {
		DPrintf("[%v]: 获取多数选票, 将会成为leader", rf.me)
		becomeLeader.Do(func() {
			DPrintf("[%v]: 在term %v 成为leader", rf.me, rf.currentTerm)
			rf.state = StateLeader
			lastLogIndex := rf.log.lastLog().Index

			// rules for leader to update nextIndex and matchIndex
			for i, _ := range rf.peers {
				rf.nextIndex[i] = lastLogIndex + 1
				rf.matchIndex[i] = 0
			}
			DPrintf("[%d]: leader nextIndex %#v", rf.me, rf.nextIndex)
			rf.appendEntries(true)
		})
	}

}

// example RequestVote RPC handler.
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (3A, 3B).

	rf.mu.Lock()

	defer rf.mu.Unlock()

	// rules2 for all servers
	if args.Term > rf.currentTerm {
		rf.setNewTerm(args.Term)
	}

	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.Grant = false
		return
	}

	lastLog := rf.log.lastLog()

	// check whether candidate has newer log then current node
	update := args.LastLogTerm > lastLog.Term || (args.LastLogTerm == lastLog.Term && args.LastLogIndex >= lastLog.Index)

	// only if current has not voted to other candidate and this candiate
	//has newer log than current node, will vote grant
	if (rf.votedFor == -1 || rf.votedFor == args.CandidateId) && update {
		reply.Grant = true
		rf.votedFor = args.CandidateId
		rf.persist()
		rf.resetElectionTimer()
		DPrintf("[%v]: term %v vote %v", rf.me, rf.currentTerm, rf.votedFor)
	} else {
		reply.Grant = false
	}

	reply.Term = rf.currentTerm
}
