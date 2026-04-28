import { motion } from 'motion/react';
import { Search, Bell, User, ChevronRight } from 'lucide-react';

export function Header({ currentView, onNavigate }) {
  return (
    <header className="fixed top-0 w-full z-50 border-b border-surface-container bg-white/80 backdrop-blur-md">
      <div className="flex justify-between items-center h-16 px-6 md:px-12 max-w-7xl mx-auto">
        <div className="flex items-center gap-8">
          <button 
            onClick={() => onNavigate('home')}
            className="text-xl font-extrabold tracking-tighter text-on-surface cursor-pointer"
          >
            MangaHub
          </button>
          <nav className="hidden md:flex items-center gap-6">
            {['Home', 'Library', 'Browse'].map((item) => (
              <button
                key={item}
                onClick={() => onNavigate(item.toLowerCase())}
                className={`text-sm font-semibold tracking-tight transition-all relative py-1 ${
                  currentView === item.toLowerCase() 
                    ? 'text-primary' 
                    : 'text-on-surface-variant hover:text-on-surface'
                }`}
              >
                {item}
                {currentView === item.toLowerCase() && (
                  <motion.div 
                    layoutId="nav-underline"
                    className="absolute bottom-0 left-0 w-full h-0.5 bg-primary"
                  />
                )}
              </button>
            ))}
          </nav>
        </div>

        <div className="flex items-center gap-4">
          <div className="relative hidden lg:block w-72">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-on-surface-variant w-4 h-4" />
            <input 
              placeholder="Search manga, authors..." 
              className="w-full bg-surface-container-low border-none rounded-full py-2 pl-10 pr-4 text-sm focus:ring-2 focus:ring-primary/20 transition-all outline-none"
            />
          </div>
          <div className="flex items-center gap-1">
            <button className="p-2 hover:bg-surface-container rounded-lg transition-all active:scale-95">
              <Bell className="w-5 h-5 text-on-surface-variant" />
            </button>
            <button 
              onClick={() => onNavigate('login')}
              className="p-2 hover:bg-surface-container rounded-lg transition-all active:scale-95"
            >
              <User className="w-5 h-5 text-on-surface-variant" />
            </button>
          </div>
        </div>
      </div>
    </header>
  );
}

export function Footer() {
  return (
    <footer className="w-full border-t border-surface-container bg-white py-12 px-6 md:px-12">
      <div className="max-w-7xl mx-auto flex flex-col md:flex-row justify-between items-center gap-8">
        <div className="text-center md:text-left">
          <h2 className="text-lg font-bold text-on-surface-variant/40 mb-2">MangaHub</h2>
          <p className="text-xs text-on-surface-variant/60">© 2024 MangaHub. All rights reserved.</p>
        </div>
        <div className="flex flex-wrap justify-center gap-8">
          {['Privacy Policy', 'Terms of Service', 'Help Center', 'Contact Us'].map(l => (
            <a key={l} href="#" className="text-xs text-on-surface-variant/60 hover:text-primary transition-all underline-offset-4 hover:underline">{l}</a>
          ))}
        </div>
      </div>
    </footer>
  );
}
