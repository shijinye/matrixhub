// Copyright The MatrixHub Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model_test

import (
	"context"

	v1alpha1model "github.com/matrixhub-ai/matrixhub/test/client/v1alpha1/model"
	v1alpha1project "github.com/matrixhub-ai/matrixhub/test/client/v1alpha1/project"
	"github.com/matrixhub-ai/matrixhub/test/tools"

	"github.com/antihax/optional"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Model", Label("model"), func() {
	var (
		ctx         context.Context
		modelsApi   *v1alpha1model.ModelsApiService
		projectsApi *v1alpha1project.ProjectsApiService
	)

	BeforeEach(func() {
		ctx = context.Background()
		modelsApi = tools.GetV1alpha1ModelsApi()
		projectsApi = tools.GetV1alpha1ProjectsApi()
	})

	// ═══════════════════════════════════════════════════════════
	// 1. Label APIs
	// ═══════════════════════════════════════════════════════════
	Context("ListModelTaskLabels API", func() {
		It("should list task labels successfully", Label("M00001"), func() {
			resp, _, err := modelsApi.ModelsListModelTaskLabels(ctx)
			Expect(err).NotTo(HaveOccurred())

			GinkgoWriter.Printf("Task labels: found %d labels\n", len(resp.Items))
		})
	})

	Context("ListModelFrameLabels API", func() {
		It("should list framework/library labels successfully", Label("M00002"), func() {
			resp, _, err := modelsApi.ModelsListModelFrameLabels(ctx)
			Expect(err).NotTo(HaveOccurred())

			GinkgoWriter.Printf("Frame labels: found %d labels\n", len(resp.Items))
		})
	})

	// ═══════════════════════════════════════════════════════════
	// 2. CRUD APIs (CreateModel / GetModel / ListModels / DeleteModel)
	// ═══════════════════════════════════════════════════════════
	Context("CRUD APIs", func() {
		var projectName string

		BeforeEach(func() {
			projectName = tools.GenerateTestProjectName("model-crud")
			_, _, err := projectsApi.ProjectsCreateProject(ctx, v1alpha1project.V1alpha1CreateProjectRequest{
				Name: projectName,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			_, _, _ = projectsApi.ProjectsDeleteProject(ctx, projectName)
		})

		// --- CreateModel ---
		It("should create a model successfully", Label("M00003"), func() {
			modelName := tools.GenerateTestModelName("model")

			_, _, err := modelsApi.ModelsCreateModel(ctx, v1alpha1model.V1alpha1CreateModelRequest{
				Project: projectName,
				Name:    modelName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Verify model is created via GetModel
			getResp, _, err := modelsApi.ModelsGetModel(ctx, projectName, modelName)
			Expect(err).NotTo(HaveOccurred())
			Expect(getResp.Name).To(Equal(modelName))
			Expect(getResp.Project).To(Equal(projectName))

			GinkgoWriter.Printf("Created model: name=%v, project=%v\n", getResp.Name, getResp.Project)
		})

		It("should fail to create duplicate model", Label("M00004"), func() {
			modelName := tools.GenerateTestModelName("dup-model")

			_, _, err := modelsApi.ModelsCreateModel(ctx, v1alpha1model.V1alpha1CreateModelRequest{
				Project: projectName,
				Name:    modelName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Try to create duplicate
			_, _, err = modelsApi.ModelsCreateModel(ctx, v1alpha1model.V1alpha1CreateModelRequest{
				Project: projectName,
				Name:    modelName,
			})
			Expect(err).To(HaveOccurred(), "duplicate model should fail")

			GinkgoWriter.Printf("Duplicate model error: %v\n", err)
		})

		It("should fail to create a model with empty project", Label("M00005"), func() {
			modelName := tools.GenerateTestModelName("model")
			_, _, err := modelsApi.ModelsCreateModel(ctx, v1alpha1model.V1alpha1CreateModelRequest{
				Project: "",
				Name:    modelName,
			})
			Expect(err).To(HaveOccurred())
		})

		It("should fail to create a model with empty name", Label("M00006"), func() {
			_, _, err := modelsApi.ModelsCreateModel(ctx, v1alpha1model.V1alpha1CreateModelRequest{
				Project: projectName,
				Name:    "",
			})
			Expect(err).To(HaveOccurred())
		})

		It("should fail to create a model in non-existent project", Label("M00007"), func() {
			modelName := tools.GenerateTestModelName("model")
			_, _, err := modelsApi.ModelsCreateModel(ctx, v1alpha1model.V1alpha1CreateModelRequest{
				Project: "non-existent-project-xyz",
				Name:    modelName,
			})
			Expect(err).To(HaveOccurred())
		})

		// --- GetModel ---
		It("should get an existing model with all expected fields", Label("M00008"), func() {
			modelName := tools.GenerateTestModelName("get-model")

			_, _, err := modelsApi.ModelsCreateModel(ctx, v1alpha1model.V1alpha1CreateModelRequest{
				Project: projectName,
				Name:    modelName,
			})
			Expect(err).NotTo(HaveOccurred())

			getResp, _, err := modelsApi.ModelsGetModel(ctx, projectName, modelName)
			Expect(err).NotTo(HaveOccurred())
			Expect(getResp.Name).To(Equal(modelName))
			Expect(getResp.Project).To(Equal(projectName))
			Expect(getResp.Id).NotTo(BeZero())
			Expect(getResp.CreatedAt).NotTo(BeEmpty())

			GinkgoWriter.Printf("GetModel: id=%v, name=%v, project=%v\n", getResp.Id, getResp.Name, getResp.Project)
		})

		It("should fail to get a non-existent model", Label("M00009"), func() {
			_, _, err := modelsApi.ModelsGetModel(ctx, projectName, "non-existent-model-xyz")
			Expect(err).To(HaveOccurred())
		})

		It("should fail to get a model from non-existent project", Label("M00010"), func() {
			_, _, err := modelsApi.ModelsGetModel(ctx, "nonexistent", "nonexistent-model")
			Expect(err).To(HaveOccurred())
		})

		// --- ListModels ---
		It("should list models with pagination", Label("M00011"), func() {
			modelName := tools.GenerateTestModelName("list-model")

			_, _, err := modelsApi.ModelsCreateModel(ctx, v1alpha1model.V1alpha1CreateModelRequest{
				Project: projectName,
				Name:    modelName,
			})
			Expect(err).NotTo(HaveOccurred())

			listResp, _, err := modelsApi.ModelsListModels(ctx, &v1alpha1model.ModelsApiModelsListModelsOpts{
				Page:     optional.NewInt32(1),
				PageSize: optional.NewInt32(50),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(listResp.Items).NotTo(BeNil())
			Expect(listResp.Pagination).NotTo(BeNil())

			GinkgoWriter.Printf("ListModels: total=%v, page=%v\n",
				listResp.Pagination.Total, listResp.Pagination.Page)
		})

		It("should filter models by project", Label("M00012"), func() {
			modelName := tools.GenerateTestModelName("filter-model")

			_, _, err := modelsApi.ModelsCreateModel(ctx, v1alpha1model.V1alpha1CreateModelRequest{
				Project: projectName,
				Name:    modelName,
			})
			Expect(err).NotTo(HaveOccurred())

			listResp, _, err := modelsApi.ModelsListModels(ctx, &v1alpha1model.ModelsApiModelsListModelsOpts{
				Project:  optional.NewString(projectName),
				Page:     optional.NewInt32(1),
				PageSize: optional.NewInt32(20),
			})
			Expect(err).NotTo(HaveOccurred())

			var found bool
			for _, m := range listResp.Items {
				if m.Name == modelName {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue(), "created model should be in the project-filtered list")
		})

		It("should search models by keyword", Label("M00013"), func() {
			modelName := tools.GenerateTestModelName("search-target")

			_, _, err := modelsApi.ModelsCreateModel(ctx, v1alpha1model.V1alpha1CreateModelRequest{
				Project: projectName,
				Name:    modelName,
			})
			Expect(err).NotTo(HaveOccurred())

			listResp, _, err := modelsApi.ModelsListModels(ctx, &v1alpha1model.ModelsApiModelsListModelsOpts{
				Project:  optional.NewString(projectName),
				Search:   optional.NewString("search-target"),
				Page:     optional.NewInt32(1),
				PageSize: optional.NewInt32(10),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(listResp.Items)).To(BeNumerically(">=", 1))
		})

		It("should respect page_size=1", Label("M00014"), func() {
			// Create 2 models
			for i := 0; i < 2; i++ {
				name := tools.GenerateTestModelName("paging")
				_, _, err := modelsApi.ModelsCreateModel(ctx, v1alpha1model.V1alpha1CreateModelRequest{
					Project: projectName,
					Name:    name,
				})
				Expect(err).NotTo(HaveOccurred())
			}

			listResp, _, err := modelsApi.ModelsListModels(ctx, &v1alpha1model.ModelsApiModelsListModelsOpts{
				Project:  optional.NewString(projectName),
				Page:     optional.NewInt32(1),
				PageSize: optional.NewInt32(1),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(listResp.Items)).To(Equal(1))
			Expect(listResp.Pagination.Total).To(BeNumerically(">=", 2))
		})

		It("should return empty list for non-existent project filter", Label("M00015"), func() {
			listResp, _, err := modelsApi.ModelsListModels(ctx, &v1alpha1model.ModelsApiModelsListModelsOpts{
				Project:  optional.NewString("nonexistent-project-xyz"),
				Page:     optional.NewInt32(1),
				PageSize: optional.NewInt32(10),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(listResp.Items)).To(Equal(0))
		})

		// --- DeleteModel ---
		It("should delete an existing model and verify", Label("M00016"), func() {
			modelName := tools.GenerateTestModelName("delete-model")

			_, _, err := modelsApi.ModelsCreateModel(ctx, v1alpha1model.V1alpha1CreateModelRequest{
				Project: projectName,
				Name:    modelName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Delete
			_, _, err = modelsApi.ModelsDeleteModel(ctx, projectName, modelName)
			Expect(err).NotTo(HaveOccurred())

			// Verify deleted via GetModel
			_, _, err = modelsApi.ModelsGetModel(ctx, projectName, modelName)
			Expect(err).To(HaveOccurred(), "get should fail after delete")
		})

		It("should fail to delete a non-existent model", Label("M00017"), func() {
			_, _, err := modelsApi.ModelsDeleteModel(ctx, projectName, "non-existent-model-xyz")
			Expect(err).To(HaveOccurred())
		})

		It("should fail to delete an already deleted model", Label("M00018"), func() {
			modelName := tools.GenerateTestModelName("double-delete")

			_, _, err := modelsApi.ModelsCreateModel(ctx, v1alpha1model.V1alpha1CreateModelRequest{
				Project: projectName,
				Name:    modelName,
			})
			Expect(err).NotTo(HaveOccurred())

			// First delete
			_, _, err = modelsApi.ModelsDeleteModel(ctx, projectName, modelName)
			Expect(err).NotTo(HaveOccurred())

			// Second delete should fail
			_, _, err = modelsApi.ModelsDeleteModel(ctx, projectName, modelName)
			Expect(err).To(HaveOccurred(), "double delete should fail")
		})

		// --- Full lifecycle ---
		It("should handle full create-get-list-delete lifecycle", Label("M00019"), func() {
			modelName := tools.GenerateTestModelName("lifecycle")

			// Create
			_, _, err := modelsApi.ModelsCreateModel(ctx, v1alpha1model.V1alpha1CreateModelRequest{
				Project: projectName,
				Name:    modelName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Get
			getResp, _, err := modelsApi.ModelsGetModel(ctx, projectName, modelName)
			Expect(err).NotTo(HaveOccurred())
			Expect(getResp.Name).To(Equal(modelName))
			Expect(getResp.Project).To(Equal(projectName))

			// List
			listResp, _, err := modelsApi.ModelsListModels(ctx, &v1alpha1model.ModelsApiModelsListModelsOpts{
				Project:  optional.NewString(projectName),
				Page:     optional.NewInt32(1),
				PageSize: optional.NewInt32(100),
			})
			Expect(err).NotTo(HaveOccurred())
			var found bool
			for _, m := range listResp.Items {
				if m.Name == modelName {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue())

			// Delete
			_, _, err = modelsApi.ModelsDeleteModel(ctx, projectName, modelName)
			Expect(err).NotTo(HaveOccurred())

			// Verify deleted
			_, _, err = modelsApi.ModelsGetModel(ctx, projectName, modelName)
			Expect(err).To(HaveOccurred())
		})
	})

	// ═══════════════════════════════════════════════════════════
	// 3. ListModelRevisions (requires pre-existing model with git data)
	// ═══════════════════════════════════════════════════════════
	Context("ListModelRevisions API", Label("git"), func() {
		var (
			gitProject string
			gitModel   string
		)

		BeforeEach(func() {
			gitProject = tools.GetGitModelProject()
			gitModel = tools.GetGitModelName()

			// Verify the git model exists, skip if not available
			_, _, err := modelsApi.ModelsGetModel(ctx, gitProject, gitModel)
			if err != nil {
				Skip("GIT_MODEL not available: " + gitProject + "/" + gitModel + " — set MATRIXHUB_GIT_PROJECT and MATRIXHUB_GIT_MODEL env vars")
			}
		})

		It("should list revisions successfully", Label("M00020"), func() {
			resp, _, err := modelsApi.ModelsListModelRevisions(ctx, gitProject, gitModel)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Items).NotTo(BeNil())

			GinkgoWriter.Printf("Revisions: branches=%d, tags=%d\n",
				len(resp.Items.Branches), len(resp.Items.Tags))
		})

		It("should fail for non-existent model", Label("M00021"), func() {
			_, _, err := modelsApi.ModelsListModelRevisions(ctx, "nonexistent", "nonexistent")
			Expect(err).To(HaveOccurred())
		})
	})

	// ═══════════════════════════════════════════════════════════
	// 4. ListModelCommits + GetModelCommit (requires git data)
	// ═══════════════════════════════════════════════════════════
	Context("Commits APIs", Label("git"), func() {
		var (
			gitProject string
			gitModel   string
		)

		BeforeEach(func() {
			gitProject = tools.GetGitModelProject()
			gitModel = tools.GetGitModelName()

			_, _, err := modelsApi.ModelsGetModel(ctx, gitProject, gitModel)
			if err != nil {
				Skip("GIT_MODEL not available: " + gitProject + "/" + gitModel)
			}
		})

		It("should list commits with default branch", Label("M00022"), func() {
			resp, _, err := modelsApi.ModelsListModelCommits(ctx, gitProject, gitModel, &v1alpha1model.ModelsApiModelsListModelCommitsOpts{
				Page:     optional.NewInt32(1),
				PageSize: optional.NewInt32(10),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Items).NotTo(BeNil())
			Expect(resp.Pagination).NotTo(BeNil())

			GinkgoWriter.Printf("Commits: count=%d, total=%d\n", len(resp.Items), resp.Pagination.Total)
		})

		It("should list commits with revision=main", Label("M00023"), func() {
			resp, _, err := modelsApi.ModelsListModelCommits(ctx, gitProject, gitModel, &v1alpha1model.ModelsApiModelsListModelCommitsOpts{
				Revision: optional.NewString("main"),
				Page:     optional.NewInt32(1),
				PageSize: optional.NewInt32(10),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Items).NotTo(BeNil())
		})

		It("should list commits with pagination", Label("M00024"), func() {
			resp, _, err := modelsApi.ModelsListModelCommits(ctx, gitProject, gitModel, &v1alpha1model.ModelsApiModelsListModelCommitsOpts{
				Page:     optional.NewInt32(1),
				PageSize: optional.NewInt32(2),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(resp.Items)).To(BeNumerically("<=", 2))
		})

		It("should fail to list commits for non-existent model", Label("M00025"), func() {
			_, _, err := modelsApi.ModelsListModelCommits(ctx, "nonexistent", "nonexistent", &v1alpha1model.ModelsApiModelsListModelCommitsOpts{
				Page:     optional.NewInt32(1),
				PageSize: optional.NewInt32(10),
			})
			Expect(err).To(HaveOccurred())
		})

		It("should get a specific commit with valid fields", Label("M00026"), func() {
			// Get first commit ID
			listResp, _, err := modelsApi.ModelsListModelCommits(ctx, gitProject, gitModel, &v1alpha1model.ModelsApiModelsListModelCommitsOpts{
				Page:     optional.NewInt32(1),
				PageSize: optional.NewInt32(1),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(listResp.Items)).To(BeNumerically(">=", 1), "need at least one commit")

			commitID := listResp.Items[0].Id
			Expect(commitID).NotTo(BeEmpty())

			// Get specific commit
			commit, _, err := modelsApi.ModelsGetModelCommit(ctx, gitProject, gitModel, commitID)
			Expect(err).NotTo(HaveOccurred())
			Expect(commit.Id).To(Equal(commitID))
			Expect(commit.Message).NotTo(BeEmpty())
			Expect(commit.AuthorName).NotTo(BeEmpty())
			Expect(commit.AuthorDate).NotTo(BeEmpty())
			// Commit ID should be 40-char SHA-1
			Expect(len(commit.Id)).To(Equal(40))

			GinkgoWriter.Printf("GetModelCommit: id=%v, author=%v, message=%v\n",
				commit.Id[:8], commit.AuthorName, commit.Message)
		})

		It("should fail to get non-existent commit ID", Label("M00027"), func() {
			_, _, err := modelsApi.ModelsGetModelCommit(ctx, gitProject, gitModel,
				"0000000000000000000000000000000000000000")
			Expect(err).To(HaveOccurred())
		})

		It("should fail to get commit from non-existent model", Label("M00028"), func() {
			_, _, err := modelsApi.ModelsGetModelCommit(ctx, "nonexistent", "nonexistent", "abc123")
			Expect(err).To(HaveOccurred())
		})
	})

	// ═══════════════════════════════════════════════════════════
	// 5. GetModelTree (requires git data)
	// ═══════════════════════════════════════════════════════════
	Context("GetModelTree API", Label("git"), func() {
		var (
			gitProject string
			gitModel   string
		)

		BeforeEach(func() {
			gitProject = tools.GetGitModelProject()
			gitModel = tools.GetGitModelName()

			_, _, err := modelsApi.ModelsGetModel(ctx, gitProject, gitModel)
			if err != nil {
				Skip("GIT_MODEL not available: " + gitProject + "/" + gitModel)
			}
		})

		It("should get root tree successfully", Label("M00029"), func() {
			resp, _, err := modelsApi.ModelsGetModelTree(ctx, gitProject, gitModel, &v1alpha1model.ModelsApiModelsGetModelTreeOpts{})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Items).NotTo(BeNil())

			GinkgoWriter.Printf("GetModelTree root: found %d entries\n", len(resp.Items))

			// Validate first item structure if available
			if len(resp.Items) > 0 {
				item := resp.Items[0]
				Expect(item.Name).NotTo(BeEmpty())
				Expect(item.Path).NotTo(BeEmpty())
				Expect(item.Type_).NotTo(BeNil())
				GinkgoWriter.Printf("  first item: name=%v, type=%v, path=%v\n", item.Name, *item.Type_, item.Path)
			}
		})

		It("should get tree with revision=main", Label("M00030"), func() {
			resp, _, err := modelsApi.ModelsGetModelTree(ctx, gitProject, gitModel, &v1alpha1model.ModelsApiModelsGetModelTreeOpts{
				Revision: optional.NewString("main"),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Items).NotTo(BeNil())
		})

		It("should handle non-existent path gracefully", Label("M00031"), func() {
			resp, _, err := modelsApi.ModelsGetModelTree(ctx, gitProject, gitModel, &v1alpha1model.ModelsApiModelsGetModelTreeOpts{
				Path: optional.NewString("nonexistent-path-xyz"),
			})
			// Either returns error or empty items
			if err == nil {
				Expect(len(resp.Items)).To(Equal(0))
			}
		})

		It("should fail for non-existent model", Label("M00032"), func() {
			_, _, err := modelsApi.ModelsGetModelTree(ctx, "nonexistent", "nonexistent", &v1alpha1model.ModelsApiModelsGetModelTreeOpts{})
			Expect(err).To(HaveOccurred())
		})
	})

	// ═══════════════════════════════════════════════════════════
	// 6. GetModelBlob (requires git data)
	// ═══════════════════════════════════════════════════════════
	Context("GetModelBlob API", Label("git"), func() {
		var (
			gitProject    string
			gitModel      string
			firstFilePath string
			firstDirPath  string
		)

		BeforeEach(func() {
			gitProject = tools.GetGitModelProject()
			gitModel = tools.GetGitModelName()

			_, _, err := modelsApi.ModelsGetModel(ctx, gitProject, gitModel)
			if err != nil {
				Skip("GIT_MODEL not available: " + gitProject + "/" + gitModel)
			}

			// Discover file and directory paths from tree
			treeResp, _, err := modelsApi.ModelsGetModelTree(ctx, gitProject, gitModel, &v1alpha1model.ModelsApiModelsGetModelTreeOpts{})
			if err != nil {
				Skip("cannot get tree for GIT_MODEL")
			}

			for _, item := range treeResp.Items {
				if item.Type_ == nil {
					continue
				}
				if *item.Type_ == v1alpha1model.FILE_V1alpha1FileType && firstFilePath == "" {
					firstFilePath = item.Path
				}
				if *item.Type_ == v1alpha1model.DIR_V1alpha1FileType && firstDirPath == "" {
					firstDirPath = item.Path
				}
			}
		})

		It("should get blob for a file with valid fields", Label("M00033"), func() {
			if firstFilePath == "" {
				Skip("no file found in tree")
			}

			resp, _, err := modelsApi.ModelsGetModelBlob(ctx, gitProject, gitModel, &v1alpha1model.ModelsApiModelsGetModelBlobOpts{
				Path: optional.NewString(firstFilePath),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Name).NotTo(BeEmpty())
			Expect(resp.Path).NotTo(BeEmpty())
			Expect(resp.Type_).NotTo(BeNil())
			Expect(*resp.Type_).To(Equal(v1alpha1model.FILE_V1alpha1FileType))

			GinkgoWriter.Printf("GetModelBlob: name=%v, path=%v, size=%v, sha256=%v\n",
				resp.Name, resp.Path, resp.Size, resp.Sha256)
		})

		It("should get blob with revision=main", Label("M00034"), func() {
			if firstFilePath == "" {
				Skip("no file found in tree")
			}

			resp, _, err := modelsApi.ModelsGetModelBlob(ctx, gitProject, gitModel, &v1alpha1model.ModelsApiModelsGetModelBlobOpts{
				Revision: optional.NewString("main"),
				Path:     optional.NewString(firstFilePath),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Name).NotTo(BeEmpty())
		})

		It("should fail for non-existent file path", Label("M00035"), func() {
			_, _, err := modelsApi.ModelsGetModelBlob(ctx, gitProject, gitModel, &v1alpha1model.ModelsApiModelsGetModelBlobOpts{
				Path: optional.NewString("nonexistent-file-xyz.txt"),
			})
			Expect(err).To(HaveOccurred())
		})

		It("should fail for directory path", Label("M00036"), func() {
			if firstDirPath == "" {
				Skip("no directory found in tree")
			}

			_, _, err := modelsApi.ModelsGetModelBlob(ctx, gitProject, gitModel, &v1alpha1model.ModelsApiModelsGetModelBlobOpts{
				Path: optional.NewString(firstDirPath),
			})
			Expect(err).To(HaveOccurred())
		})

		It("should fail for non-existent model", Label("M00037"), func() {
			_, _, err := modelsApi.ModelsGetModelBlob(ctx, "nonexistent", "nonexistent", &v1alpha1model.ModelsApiModelsGetModelBlobOpts{
				Path: optional.NewString("README.md"),
			})
			Expect(err).To(HaveOccurred())
		})

		It("should fail for empty path", Label("M00038"), func() {
			_, _, err := modelsApi.ModelsGetModelBlob(ctx, gitProject, gitModel, &v1alpha1model.ModelsApiModelsGetModelBlobOpts{})
			Expect(err).To(HaveOccurred())
		})
	})
})
