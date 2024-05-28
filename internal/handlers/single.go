// --------------------------------------------------------------------------------
// Author: Thomas F McGeehan V
//
// This file is part of a software project developed by Thomas F McGeehan V.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// For more information about the MIT License, please visit:
// https://opensource.org/licenses/MIT
//
// Acknowledgment appreciated but not required.
// --------------------------------------------------------------------------------

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/TFMV/FuzzyMatchFinder/internal/matcher"
	"github.com/jackc/pgx/v5/pgxpool"
)

func MatchSingleHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req matcher.MatchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Insert the single record into the database with a unique run_id
		runID := matcher.CreateNewRun(pool, "Single Record Matching")
		req.RunID = runID
		fmt.Println("runID: ", runID)
		// Clear existing entries for this run_id
		matcher.ClearOldCandidates(pool, runID)

		// Process the single record
		matcher.ProcessSingleRecord(pool, req)

		// Generate TF/IDF vectors for the single record
		matcher.GenerateTFIDF(pool, runID)

		// Insert vector embeddings using Python script
		scriptPath := "./python-ml/generate_embeddings.py"
		if err := matcher.GenerateEmbeddingsPythonScript(scriptPath, runID); err != nil {
			http.Error(w, fmt.Sprintf("Failed to generate embeddings: %v", err), http.StatusInternalServerError)
			return
		}

		// Find matches
		candidates := matcher.FindMatches(req, matcher.NewScorer(), pool)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(candidates)
	}
}
