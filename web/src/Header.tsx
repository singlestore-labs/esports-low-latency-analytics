import React from 'react';
import { Link } from 'react-router-dom';
import { WithChildren } from './util';
import { HomeIcon } from '@heroicons/react/solid';

type Props = WithChildren;

const Header: React.FC<Props> = ({ children }: Props) => (
    <header className="sticky top-0 z-50 flex justify-between items-center border-b-2 border-gray-100 py-6 mb-3 bg-white font-medium">
        <Link to="/" className="text-gray-500 hover:text-indigo-500">
            <HomeIcon className="h-6 px-4" />
        </Link>
        {children}
    </header>
);

export default Header;
