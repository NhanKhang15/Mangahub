import { motion } from 'motion/react';
import { BookOpen, Mail, Lock, Eye, Github, Chrome } from 'lucide-react';

export function Login() {
  return (
    <div className="flex items-center justify-center py-20">
      <motion.div 
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        className="w-full max-w-[440px] bg-white rounded-3xl shadow-xl shadow-primary/5 p-8 md:p-12 border border-surface-container"
      >
        <div className="flex flex-col items-center mb-10 text-center">
          <div className="w-16 h-16 bg-primary/10 rounded-2xl flex items-center justify-center mb-6">
            <BookOpen className="w-8 h-8 text-primary" />
          </div>
          <h1 className="text-3xl font-extrabold text-on-surface mb-2 tracking-tight">Welcome back</h1>
          <p className="text-on-surface-variant font-medium">Continue your reading journey where you left off.</p>
        </div>

        <form className="space-y-5" onSubmit={(e) => e.preventDefault()}>
          <div className="space-y-1.5">
            <label className="text-xs font-bold text-on-surface ml-1 uppercase tracking-widest text-on-surface-variant/50">Email Address</label>
            <div className="relative">
              <Mail className="absolute left-4 top-1/2 -translate-y-1/2 text-on-surface-variant/40 w-5 h-5" />
              <input className="w-full pl-12 pr-4 py-4 bg-surface-container-low border-none rounded-2xl focus:ring-2 focus:ring-primary/20 focus:bg-white transition-all outline-none font-medium" placeholder="name@example.com" type="email" />
            </div>
          </div>

          <div className="space-y-1.5">
            <div className="flex justify-between items-center ml-1">
              <label className="text-xs font-bold text-on-surface-variant/50 uppercase tracking-widest">Password</label>
              <button className="text-xs font-bold text-primary hover:underline">Forgot?</button>
            </div>
            <div className="relative">
              <Lock className="absolute left-4 top-1/2 -translate-y-1/2 text-on-surface-variant/40 w-5 h-5" />
              <input className="w-full pl-12 pr-12 py-4 bg-surface-container-low border-none rounded-2xl focus:ring-2 focus:ring-primary/20 focus:bg-white transition-all outline-none font-medium" placeholder="••••••••" type="password" />
              <button className="absolute right-4 top-1/2 -translate-y-1/2 text-on-surface-variant/40 hover:text-on-surface">
                <Eye className="w-5 h-5" />
              </button>
            </div>
          </div>

          <div className="flex items-center gap-2 py-1 ml-1">
            <input type="checkbox" className="w-4 h-4 rounded border-surface-container text-primary focus:ring-primary/20" id="remember" />
            <label htmlFor="remember" className="text-xs font-bold text-on-surface-variant">Remember me for 30 days</label>
          </div>

          <button className="w-full bg-primary text-white py-4 rounded-2xl font-bold shadow-lg shadow-primary/20 hover:bg-primary-container active:scale-[0.98] transition-all">
            Sign in to MangaHub
          </button>
        </form>

        <div className="relative my-10">
          <div className="absolute inset-0 flex items-center"><div className="w-full border-t border-surface-container" /></div>
          <div className="relative flex justify-center text-[10px] font-extrabold text-on-surface-variant/40 tracking-[0.2em] uppercase"><span className="bg-white px-4">Or continue with</span></div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <button className="flex items-center justify-center gap-2 py-3.5 bg-surface-container-low hover:bg-surface-container border border-surface-container/50 rounded-2xl transition-all font-bold text-sm">
            <Chrome className="w-4 h-4" /> Google
          </button>
          <button className="flex items-center justify-center gap-2 py-3.5 bg-surface-container-low hover:bg-surface-container border border-surface-container/50 rounded-2xl transition-all font-bold text-sm">
            <Github className="w-4 h-4" /> Github
          </button>
        </div>

        <p className="mt-10 text-center text-sm font-medium text-on-surface-variant">
          Don't have an account? <button className="text-primary font-bold hover:underline ml-1">Create an account</button>
        </p>
      </motion.div>
    </div>
  );
}
