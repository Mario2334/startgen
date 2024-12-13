const Post = require('./models/Post');

const resolvers = {
  Query: {
    getPosts: async () => await Post.find(),
    getPost: async (_, { id }) => await Post.findById(id),
  },
  Mutation: {
    addPost: async (_, { title, content }) => {
      const post = new Post({ title, content });
      return await post.save();
    },
    updatePost: async (_, { id, title, content }) => {
      return await Post.findByIdAndUpdate(
        id,
        { title, content },
        { new: true }
      );
    },
    deletePost: async (_, { id }) => {
      await Post.findByIdAndDelete(id);
      return 'Post deleted';
    },
  },
};

module.exports = resolvers;